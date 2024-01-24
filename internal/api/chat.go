// Copyright © 2023 OpenIM open source community. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"context"
	"fmt"
	"github.com/BioforestChain/dweb-browser-im-chats/pkg/common/apicall"
	"github.com/BioforestChain/dweb-browser-im-chats/pkg/common/apistruct"
	constant2 "github.com/BioforestChain/dweb-browser-im-chats/pkg/common/constant"
	"github.com/BioforestChain/dweb-browser-im-chats/pkg/common/db/cache"
	"github.com/BioforestChain/dweb-browser-im-chats/pkg/common/mctx"
	"github.com/BioforestChain/dweb-browser-im-chats/pkg/common/sign"
	"github.com/BioforestChain/dweb-browser-im-chats/pkg/common/util/hash"
	"github.com/BioforestChain/dweb-browser-im-chats/pkg/common/util/number"
	"github.com/OpenIMSDK/protocol/sdkws"
	"github.com/OpenIMSDK/tools/checker"
	"github.com/OpenIMSDK/tools/log"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/tools/a2r"
	"github.com/OpenIMSDK/tools/apiresp"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"github.com/BioforestChain/dweb-browser-im-chats/pkg/common/config"
	"github.com/BioforestChain/dweb-browser-im-chats/pkg/proto/admin"
	"github.com/BioforestChain/dweb-browser-im-chats/pkg/proto/chat"
)

func NewChat(chatConn, adminConn grpc.ClientConnInterface) *ChatApi {
	return &ChatApi{chatClient: chat.NewChatClient(chatConn), adminClient: admin.NewAdminClient(adminConn), imApiCaller: apicall.NewCallerInterface()}
}

type ChatApi struct {
	chatClient  chat.ChatClient
	adminClient admin.AdminClient
	imApiCaller apicall.CallerInterface
}

// ################## ACCOUNT ##################

func (o *ChatApi) SendVerifyCode(c *gin.Context) {
	req := chat.SendVerifyCodeReq{}

	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, err)
		return
	}
	ip, err := o.getClientIP(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	req.Ip = ip
	resp, err := o.chatClient.SendVerifyCode(c, &req)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, resp)
}

func (o *ChatApi) VerifyCode(c *gin.Context) {
	a2r.Call(chat.ChatClient.VerifyCode, o.chatClient, c)
}

// RegisterUser
//
//	@Description: 原系统注册流程，注册用户
//	@receiver o
//	@param c
func (o *ChatApi) RegisterUser(c *gin.Context) {
	var (
		req  chat.RegisterUserReq
		resp apistruct.UserRegisterResp
	)
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, err)
		return
	}
	log.ZInfo(c, "registerUser", "req", &req)
	if err := checker.Validate(&req); err != nil {
		apiresp.GinError(c, err) // 参数校验失败
		return
	}
	ip, err := o.getClientIP(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	req.Ip = ip
	respRegisterUser, err := o.chatClient.RegisterUser(c, &req)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	userInfo := &sdkws.UserInfo{
		UserID:     respRegisterUser.UserID,
		Nickname:   req.User.Nickname,
		FaceURL:    req.User.FaceURL,
		CreateTime: time.Now().UnixMilli(),
	}
	err = o.imApiCaller.RegisterUser(c, []*sdkws.UserInfo{userInfo})
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	imToken, err := o.imApiCaller.ImAdminTokenWithDefaultAdmin(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiCtx := mctx.WithApiToken(c, imToken)
	rpcCtx := mctx.WithAdminUser(c)
	if resp, err := o.adminClient.FindDefaultFriend(rpcCtx, &admin.FindDefaultFriendReq{}); err == nil {
		_ = o.imApiCaller.ImportFriend(apiCtx, respRegisterUser.UserID, resp.UserIDs)
	}
	if resp, err := o.adminClient.FindDefaultGroup(rpcCtx, &admin.FindDefaultGroupReq{}); err == nil {
		_ = o.imApiCaller.InviteToGroup(apiCtx, respRegisterUser.UserID, resp.GroupIDs)
	}
	if req.AutoLogin {
		resp.ImToken, err = o.imApiCaller.UserToken(c, respRegisterUser.UserID, req.Platform)
		if err != nil {
			apiresp.GinError(c, err)
			return
		}
	}
	resp.ChatToken = respRegisterUser.ChatToken
	resp.UserID = respRegisterUser.UserID
	log.ZInfo(c, "registerUser api", "resp", &resp)
	apiresp.GinSuccess(c, &resp)
}

// Challenge
//
//	@Description: 下发给前端一个验证码（挑战,随机6位数字）
//	@receiver o
//	@param c
func (o *ChatApi) Challenge(c *gin.Context) {
	var (
		req  chat.ChallengeReq
		resp apistruct.ChallengeResp
	)
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, err)
		return
	}
	log.ZInfo(c, "Challenge", "req", &req)
	if req.PublicKey == "" {
		apiresp.GinError(c, errs.ErrArgs.Wrap("publicKey field is required")) // 参数校验失败
		return
	}
	if err := checker.Validate(&req); err != nil {
		apiresp.GinError(c, err) // 参数校验失败
		return
	}
	publicKey := req.PublicKey
	publicKeyMd5 := hash.Md5(publicKey)
	//challenge := number.GenerateTraceId()
	challenge := number.GetRandNum(constant2.LenRandomNum)
	challengeStr := strconv.Itoa(challenge)
	rdb, err := cache.NewRedis()
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	ca := cache.NewTokenInterface(rdb)
	ca.AddTraceId(context.Background(), publicKeyMd5, challengeStr)

	ca.GetTraceId(context.Background(), publicKeyMd5)

	resp.Challenge = challengeStr
	apiresp.GinSuccess(c, resp)
}

// Auth
//
//	@Description: 为前端进行验签操作，并下发token
//	@receiver o
//	@param c
func (o *ChatApi) Auth(c *gin.Context) {
	var (
		req    chat.AuthReq
		resp   apistruct.AuthResp
		reqNew chat.RegisterUserReq
	)
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, err)
		return
	}
	log.ZInfo(c, "Auth", "req", &req)

	if err := checker.Validate(&req); err != nil {
		apiresp.GinError(c, err) // 参数校验失败
		return
	}

	signature := req.Sign
	publicKeyStr := req.PublicKey
	address := req.Address

	publicKey := req.PublicKey
	publicKeyMd5 := hash.Md5(publicKey)
	rdb, err := cache.NewRedis()
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	ca := cache.NewTokenInterface(rdb)
	challenge, err := ca.GetTraceId(context.Background(), publicKeyMd5)
	if challenge == "" {
		apiresp.GinError(c, errs.ErrArgs.Wrap("challenge void")) // 参数校验失败
		return
	}
	if address == "" {
		apiresp.GinError(c, errs.ErrArgs.Wrap("address void")) // 参数校验失败
		return
	}
	verSign := sign.VerifySign(challenge, signature, publicKeyStr)
	if !verSign {
		apiresp.GinError(c, errs.ErrArgs.Wrap("sign validation failed error "))
		return
	}

	// insert into db (gorm )
	// 1. address 明文
	// 2. pub_key
	// 3. token imToken
	reqNew.VerifyCode = "666666"
	reqNew.Platform = constant.WebPlatformID
	reqNew.AutoLogin = true
	timeStamp := fmt.Sprintf("%d", time.Now().Unix())

	ip, err := o.getClientIP(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	reqNew.Ip = ip

	reqNew.User = &chat.RegisterUserInfo{
		Nickname: timeStamp,
		//PhoneNumber: "19900000000",
		PhoneNumber: timeStamp + "0",
		Address:     address,
		PublicKey:   publicKeyStr,
		FaceURL:     "",
		AreaCode:    "+86",
		Password:    "df10ef8509dc176d733d59549e7dbfaf",
	}

	userAccount, err := o.chatClient.GetUserByAddress(c, &chat.GetUserReq{Address: req.Address})
	if err != nil && !errs.ErrRecordNotFound.Is(err) {
		apiresp.GinError(c, errs.ErrArgs.Wrap("user account record not found")) // 账号不存在参数校验失败
		return
	}

	// 老用户
	if userAccount != nil {
		imToken, err := o.imApiCaller.UserToken(c, userAccount.UserAccount.UserID, constant.WebPlatformID)
		if err != nil {
			apiresp.GinError(c, err)
			return
		}

		chatTokenRes, err := o.adminClient.CreateToken(c, &admin.CreateTokenReq{UserID: userAccount.UserAccount.UserID, UserType: constant2.NormalUser})
		if err != nil {
			apiresp.GinError(c, err)
			return
		}

		resp.ChatToken = chatTokenRes.Token
		resp.UserID = userAccount.UserAccount.UserID
		resp.ImToken = imToken
		apiresp.GinSuccess(c, resp)
		return
	}

	respRegisterUser, err := o.chatClient.RegisterUser(c, &reqNew)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	userInfo := &sdkws.UserInfo{
		UserID:     respRegisterUser.UserID,
		Nickname:   timeStamp,
		FaceURL:    reqNew.User.FaceURL,
		CreateTime: time.Now().UnixMilli(),
	}
	err = o.imApiCaller.RegisterUser(c, []*sdkws.UserInfo{userInfo})
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	// 目前根据旧账号系统登录后，使用 Header Token: imToken
	resp.ChatToken = respRegisterUser.ChatToken
	resp.UserID = respRegisterUser.UserID

	imToken, err := o.imApiCaller.UserToken(c, respRegisterUser.UserID, constant.WebPlatformID)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}

	resp.ImToken = imToken
	log.ZInfo(c, "Auth resp: ", "resp", resp)
	apiresp.GinSuccess(c, resp)
}
func (o *ChatApi) Login(c *gin.Context) {
	var (
		req  chat.LoginReq
		resp apistruct.LoginResp
	)
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, err)
		return
	}
	if err := checker.Validate(&req); err != nil {
		apiresp.GinError(c, err) // 参数校验失败
		return
	}
	ip, err := o.getClientIP(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	req.Ip = ip
	resp1, err := o.chatClient.Login(c, &req)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	imToken, err := o.imApiCaller.UserToken(c, resp1.UserID, req.Platform)

	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	resp.ImToken = imToken
	resp.UserID = resp1.UserID
	resp.ChatToken = resp1.ChatToken
	apiresp.GinSuccess(c, resp)
}

func (o *ChatApi) ResetPassword(c *gin.Context) {
	a2r.Call(chat.ChatClient.ResetPassword, o.chatClient, c)
}

func (o *ChatApi) ChangePassword(c *gin.Context) {
	a2r.Call(chat.ChatClient.ChangePassword, o.chatClient, c)
}

// ################## USER ##################

func (o *ChatApi) UpdateUserInfo(c *gin.Context) {
	var (
		req  chat.UpdateUserInfoReq
		resp apistruct.UpdateUserInfoResp
	)
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, err)
		return
	}
	log.ZInfo(c, "updateUserInfo", "req", &req)
	if err := checker.Validate(&req); err != nil {
		apiresp.GinError(c, err) // 参数校验失败
		return
	}
	respUpdate, err := o.chatClient.UpdateUserInfo(c, &req)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	opUserType, err := mctx.GetUserType(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	var imToken string
	if opUserType == constant2.NormalUser {
		imToken, err = o.imApiCaller.ImAdminTokenWithDefaultAdmin(c)
	} else if opUserType == constant2.AdminUser {
		imToken, err = o.imApiCaller.UserToken(c, config.GetIMAdmin(mctx.GetOpUserID(c)), constant.AdminPlatformID)
	} else {
		apiresp.GinError(c, errs.ErrArgs.Wrap("opUserType unknown"))
		return
	}
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	var (
		nickName string
		faceURL  string
	)
	if req.Nickname != nil {
		nickName = req.Nickname.Value
	} else {
		nickName = respUpdate.NickName
	}
	if req.FaceURL != nil {
		faceURL = req.FaceURL.Value
	} else {
		faceURL = respUpdate.FaceUrl
	}
	err = o.imApiCaller.UpdateUserInfo(mctx.WithApiToken(c, imToken), req.UserID, nickName, faceURL)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, resp)
}

func (o *ChatApi) FindUserPublicInfo(c *gin.Context) {
	a2r.Call(chat.ChatClient.FindUserPublicInfo, o.chatClient, c)
}

func (o *ChatApi) FindUserFullInfo(c *gin.Context) {
	a2r.Call(chat.ChatClient.FindUserFullInfo, o.chatClient, c)
}

//func (o *ChatApi) GetUsersFullInfo(c *gin.Context) {
//	a2r.Call(chat.ChatClient.GetUsersFullInfo, o.chatClient, c)
//}

func (o *ChatApi) SearchUserFullInfo(c *gin.Context) {
	a2r.Call(chat.ChatClient.SearchUserFullInfo, o.chatClient, c)
}

func (o *ChatApi) SearchUserPublicInfo(c *gin.Context) {
	a2r.Call(chat.ChatClient.SearchUserPublicInfo, o.chatClient, c)
}

// ################## APPLET ##################

func (o *ChatApi) FindApplet(c *gin.Context) {
	a2r.Call(admin.AdminClient.FindApplet, o.adminClient, c)
}

// ################## CONFIG ##################

func (o *ChatApi) GetClientConfig(c *gin.Context) {
	a2r.Call(admin.AdminClient.GetClientConfig, o.adminClient, c)
}

// ################## CALLBACK ##################

func (o *ChatApi) OpenIMCallback(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	req := &chat.OpenIMCallbackReq{
		Command: c.Query(constant.CallbackCommand),
		Body:    string(body),
	}
	if _, err := o.chatClient.OpenIMCallback(c, req); err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, nil)
}

func (o *ChatApi) getClientIP(c *gin.Context) (string, error) {
	if config.Config.ProxyHeader == "" {
		ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		return ip, err
	}
	ip := c.Request.Header.Get(config.Config.ProxyHeader)
	if ip == "" {
		return "", errs.ErrInternalServer.Wrap()
	}
	if ip := net.ParseIP(ip); ip == nil {
		return "", errs.ErrInternalServer.Wrap(fmt.Sprintf("parse proxy ip header %s failed", ip))
	}
	return ip, nil
}

func (o *ChatApi) UploadLogs(c *gin.Context) {
	a2r.Call(chat.ChatClient.UploadLogs, o.chatClient, c)
}

func (o *ChatApi) SearchFriend(c *gin.Context) {
	var req struct {
		UserID string `json:"userID"`
		chat.SearchUserInfoReq
	}
	if err := c.BindJSON(&req); err != nil {
		apiresp.GinError(c, err)
		return
	}
	if req.UserID == "" {
		req.UserID = mctx.GetOpUserID(c)
	}
	imToken, err := o.imApiCaller.ImAdminTokenWithDefaultAdmin(c)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	userIDs, err := o.imApiCaller.FriendUserIDs(mctx.WithApiToken(c, imToken), req.UserID)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	if len(userIDs) == 0 {
		apiresp.GinSuccess(c, &chat.SearchUserInfoResp{})
		return
	}
	req.SearchUserInfoReq.UserIDs = userIDs
	resp, err := o.chatClient.SearchUserInfo(c, &req.SearchUserInfoReq)
	if err != nil {
		apiresp.GinError(c, err)
		return
	}
	apiresp.GinSuccess(c, resp)
}
