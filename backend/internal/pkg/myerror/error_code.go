package myerror

const (
	SuccessMessage = "ok"
)

const (
	ErrorCodeOK = 1000 // 成功

	ErrorCodeJsonFormatInvalid = 1001 // JSON 格式有误

	ErrorCodeTokenUploadFailed = 1002 // 令牌上传失败

	ErrorCodeProviderRegisterFailed = 1003 // Provider 注册失败

	ErrorCodeParameterInvalid = 1004 // 参数无效

	ErrorCodeProviderDeleteFailed = 1005 // Provider 删除失败

	ErrorCodeProviderUpdateFailed = 1006 // Provider 更新失败

	ErrorCodeProviderGetFailed = 1007 // Provider 获取失败

	ErrorCodeProviderListFailed = 1008 // Provider 列表获取失败

	ErrorCodeQuotaQueryFailed = 1009 // Quota 查询失败

	ErrorCodeQuotaSyncFailed = 1010 // Quota 同步失败

	ErrorCodeMountCreateFailed = 1011 // Mount 创建失败

	ErrorCodeMountGetFailed = 1012 // Mount 获取失败

	ErrorCodeMountQuotaSyncFailed = 1013 // Mount Quota 同步失败

	ErrorCodeMountValidationFailed = 1014 // Mount 参数校验失败

	ErrorCodeQuotaResolveFailed = 1015 // Quota 策略解析失败

	ErrorCodeLoginFailed = 1016 // 登录失败

	ErrorCodeGetDriversFailed = 1017 // 获取驱动列表失败

	ErrorCodeGetDriverInfoFailed = 1018 // 获取驱动信息失败

	ErrorCodeGetFilesFailed = 1019 // 获取文件列表失败

	ErrorCodeGetFileInfoFailed = 1020 // 获取文件信息失败

	ErrorCodeDownloadResolveFailed = 1021 // 解析直链失败

	ErrorCodeDownloadCreateFailed = 1022 // 创建下载任务失败

	ErrorCodeDownloadGetFailed = 1023 // 获取下载任务失败

	ErrorCodeResetFailed = 1024 // 重置用户数据失败

	ErrorCodeGetUserInfoFailed = 1025 // 获取用户信息失败

	ErrorCodeMountDeleteFailed = 1026 // Mount 删除失败
	ErrorCodeMountUpdateFailed = 1027 // Mount 更新失败

	ErrorCodeForbidden            = 1028 // 权限不足
	ErrorCodeDownloadOpenFailed   = 1029 // 打开下载文件失败
	ErrorCodeSettingsGetFailed    = 1030 // 获取系统配置失败
	ErrorCodeSettingsUpdateFailed = 1031 // 更新系统配置失败
	ErrorCodeBackupFailed         = 1032 // 备份用户数据失败
	ErrorCodeRestoreFailed        = 1033 // 还原用户数据失败
)
