# Details

Date : 2026-06-08 04:44:21

Directory f:\\OpenBridge\\Monorepo

Total : 113 files,  14962 codes, 391 comments, 3017 blanks, all 18370 lines

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)

## Files
| filename | language | code | comment | blank | total |
| :--- | :--- | ---: | ---: | ---: | ---: |
| [README.md](/README.md) | Markdown | 16 | 0 | 10 | 26 |
| [backend/go.mod](/backend/go.mod) | XML | 44 | 0 | 4 | 48 |
| [backend/internal/config/config.go](/backend/internal/config/config.go) | Go | 68 | 3 | 11 | 82 |
| [backend/internal/config/env.go](/backend/internal/config/env.go) | Go | 62 | 0 | 15 | 77 |
| [backend/internal/domain/entity/downtask.go](/backend/internal/domain/entity/downtask.go) | Go | 21 | 4 | 3 | 28 |
| [backend/internal/domain/entity/mount\_point.go](/backend/internal/domain/entity/mount_point.go) | Go | 18 | 0 | 3 | 21 |
| [backend/internal/domain/entity/provider\_account.go](/backend/internal/domain/entity/provider_account.go) | Go | 22 | 0 | 3 | 25 |
| [backend/internal/domain/entity/quota.go](/backend/internal/domain/entity/quota.go) | Go | 9 | 0 | 2 | 11 |
| [backend/internal/domain/entity/quota\_mode.go](/backend/internal/domain/entity/quota_mode.go) | Go | 16 | 0 | 5 | 21 |
| [backend/internal/domain/entity/quota\_snapshot.go](/backend/internal/domain/entity/quota_snapshot.go) | Go | 16 | 0 | 3 | 19 |
| [backend/internal/domain/entity/rclone\_profile.go](/backend/internal/domain/entity/rclone_profile.go) | Go | 16 | 0 | 3 | 19 |
| [backend/internal/domain/entity/token.go](/backend/internal/domain/entity/token.go) | Go | 16 | 1 | 3 | 20 |
| [backend/internal/domain/interfaces/provider\_interface.go](/backend/internal/domain/interfaces/provider_interface.go) | Go | 11 | 0 | 2 | 13 |
| [backend/internal/domain/providers/baidu\_provider.go](/backend/internal/domain/providers/baidu_provider.go) | Go | 159 | 27 | 33 | 219 |
| [backend/internal/domain/providers/local\_linux\_provider.go](/backend/internal/domain/providers/local_linux_provider.go) | Go | 73 | 22 | 20 | 115 |
| [backend/internal/domain/providers/local\_windows\_provider.go](/backend/internal/domain/providers/local_windows_provider.go) | Go | 75 | 2 | 16 | 93 |
| [backend/internal/domain/providers/mock\_provider.go](/backend/internal/domain/providers/mock_provider.go) | Go | 24 | 0 | 6 | 30 |
| [backend/internal/domain/providers/quark\_provider.go](/backend/internal/domain/providers/quark_provider.go) | Go | 109 | 2 | 22 | 133 |
| [backend/internal/handler/download\_handler.go](/backend/internal/handler/download_handler.go) | Go | 195 | 0 | 34 | 229 |
| [backend/internal/handler/mount\_handler.go](/backend/internal/handler/mount_handler.go) | Go | 163 | 6 | 24 | 193 |
| [backend/internal/handler/provider\_handler.go](/backend/internal/handler/provider_handler.go) | Go | 98 | 10 | 23 | 131 |
| [backend/internal/handler/rclone\_handler.go](/backend/internal/handler/rclone_handler.go) | Go | 97 | 0 | 17 | 114 |
| [backend/internal/handler/settings\_handler.go](/backend/internal/handler/settings_handler.go) | Go | 87 | 0 | 19 | 106 |
| [backend/internal/handler/storage\_handler.go](/backend/internal/handler/storage_handler.go) | Go | 62 | 16 | 19 | 97 |
| [backend/internal/handler/system\_handler.go](/backend/internal/handler/system_handler.go) | Go | 37 | 0 | 11 | 48 |
| [backend/internal/handler/user\_handler.go](/backend/internal/handler/user_handler.go) | Go | 62 | 6 | 17 | 85 |
| [backend/internal/handler/webdav\_proxy\_handler.go](/backend/internal/handler/webdav_proxy_handler.go) | Go | 266 | 0 | 39 | 305 |
| [backend/internal/middleware/access\_log.go](/backend/internal/middleware/access_log.go) | Go | 30 | 11 | 10 | 51 |
| [backend/internal/middleware/admin.go](/backend/internal/middleware/admin.go) | Go | 108 | 11 | 18 | 137 |
| [backend/internal/middleware/request\_id.go](/backend/internal/middleware/request_id.go) | Go | 39 | 24 | 10 | 73 |
| [backend/internal/pkg/logger/logger.go](/backend/internal/pkg/logger/logger.go) | Go | 43 | 24 | 13 | 80 |
| [backend/internal/pkg/myerror/error\_code.go](/backend/internal/pkg/myerror/error_code.go) | Go | 38 | 0 | 30 | 68 |
| [backend/internal/repository/download\_repo.go](/backend/internal/repository/download_repo.go) | Go | 24 | 0 | 8 | 32 |
| [backend/internal/repository/mount\_repo.go](/backend/internal/repository/mount_repo.go) | Go | 65 | 1 | 16 | 82 |
| [backend/internal/repository/provider\_repo.go](/backend/internal/repository/provider_repo.go) | Go | 83 | 8 | 18 | 109 |
| [backend/internal/repository/quota\_repo.go](/backend/internal/repository/quota_repo.go) | Go | 14 | 3 | 6 | 23 |
| [backend/internal/repository/rclone\_profile\_repo.go](/backend/internal/repository/rclone_profile_repo.go) | Go | 34 | 0 | 10 | 44 |
| [backend/internal/tool/aria2\_client.go](/backend/internal/tool/aria2_client.go) | Go | 153 | 1 | 28 | 182 |
| [backend/internal/tool/httpresult.go](/backend/internal/tool/httpresult.go) | Go | 14 | 0 | 3 | 17 |
| [backend/internal/tool/path\_picker.go](/backend/internal/tool/path_picker.go) | Go | 190 | 0 | 24 | 214 |
| [backend/internal/tool/provider\_register.go](/backend/internal/tool/provider_register.go) | Go | 35 | 1 | 9 | 45 |
| [backend/internal/tool/secret\_codec.go](/backend/internal/tool/secret_codec.go) | Go | 56 | 0 | 12 | 68 |
| [backend/internal/usecase/download\_usecase.go](/backend/internal/usecase/download_usecase.go) | Go | 279 | 10 | 41 | 330 |
| [backend/internal/usecase/mount\_usecase.go](/backend/internal/usecase/mount_usecase.go) | Go | 555 | 19 | 52 | 626 |
| [backend/internal/usecase/provider\_usecase.go](/backend/internal/usecase/provider_usecase.go) | Go | 108 | 16 | 24 | 148 |
| [backend/internal/usecase/rclone\_usecase.go](/backend/internal/usecase/rclone_usecase.go) | Go | 547 | 0 | 73 | 620 |
| [backend/internal/usecase/settings\_usecase.go](/backend/internal/usecase/settings_usecase.go) | Go | 105 | 0 | 24 | 129 |
| [backend/internal/usecase/storage\_usecase.go](/backend/internal/usecase/storage_usecase.go) | Go | 449 | 40 | 92 | 581 |
| [backend/internal/usecase/system\_usecase.go](/backend/internal/usecase/system_usecase.go) | Go | 169 | 0 | 25 | 194 |
| [backend/internal/usecase/user\_usecase.go](/backend/internal/usecase/user_usecase.go) | Go | 246 | 22 | 48 | 316 |
| [backend/main.go](/backend/main.go) | Go | 211 | 17 | 35 | 263 |
| [backend/web/embed.go](/backend/web/embed.go) | Go | 3 | 1 | 2 | 6 |
| [build.ps1](/build.ps1) | PowerShell | 15 | 0 | 3 | 18 |
| [docs/API.md](/docs/API.md) | Markdown | 205 | 0 | 133 | 338 |
| [docs/backend\_dev.md](/docs/backend_dev.md) | Markdown | 69 | 0 | 31 | 100 |
| [docs/frontend\_dev.md](/docs/frontend_dev.md) | Markdown | 828 | 0 | 337 | 1,165 |
| [docs/homework2.md](/docs/homework2.md) | Markdown | 39 | 0 | 19 | 58 |
| [docs/openlist\_api.md](/docs/openlist_api.md) | Markdown | 135 | 0 | 45 | 180 |
| [docs/task/week1.md](/docs/task/week1.md) | Markdown | 3 | 0 | 1 | 4 |
| [docs/task/week10.md](/docs/task/week10.md) | Markdown | 1,009 | 0 | 366 | 1,375 |
| [docs/task/week11.md](/docs/task/week11.md) | Markdown | 4 | 0 | 6 | 10 |
| [docs/task/week2.md](/docs/task/week2.md) | Markdown | 246 | 0 | 95 | 341 |
| [docs/task/week3.md](/docs/task/week3.md) | Markdown | 258 | 0 | 86 | 344 |
| [docs/task/week4.md](/docs/task/week4.md) | Markdown | 180 | 0 | 99 | 279 |
| [docs/task/week5.md](/docs/task/week5.md) | Markdown | 145 | 0 | 66 | 211 |
| [docs/task/week6&7.md](/docs/task/week6&7.md) | Markdown | 128 | 0 | 49 | 177 |
| [docs/task/week8&9.md](/docs/task/week8&9.md) | Markdown | 151 | 0 | 76 | 227 |
| [docs/团队开发.md](/docs/%E5%9B%A2%E9%98%9F%E5%BC%80%E5%8F%91.md) | Markdown | 232 | 0 | 110 | 342 |
| [docs/文档框架.md](/docs/%E6%96%87%E6%A1%A3%E6%A1%86%E6%9E%B6.md) | Markdown | 24 | 0 | 13 | 37 |
| [docs/项目架构文档.md](/docs/%E9%A1%B9%E7%9B%AE%E6%9E%B6%E6%9E%84%E6%96%87%E6%A1%A3.md) | Markdown | 802 | 0 | 161 | 963 |
| [env.md](/env.md) | Markdown | 25 | 0 | 8 | 33 |
| [frontend/index.html](/frontend/index.html) | HTML | 12 | 0 | 1 | 13 |
| [frontend/package-lock.json](/frontend/package-lock.json) | JSON | 2,031 | 0 | 1 | 2,032 |
| [frontend/package.json](/frontend/package.json) | JSON | 25 | 0 | 1 | 26 |
| [frontend/src/api/endpoints.ts](/frontend/src/api/endpoints.ts) | TypeScript | 38 | 18 | 10 | 66 |
| [frontend/src/api/mount.ts](/frontend/src/api/mount.ts) | TypeScript | 22 | 6 | 7 | 35 |
| [frontend/src/api/provider.ts](/frontend/src/api/provider.ts) | TypeScript | 19 | 5 | 5 | 29 |
| [frontend/src/api/quota.ts](/frontend/src/api/quota.ts) | TypeScript | 10 | 2 | 2 | 14 |
| [frontend/src/api/rclone.ts](/frontend/src/api/rclone.ts) | TypeScript | 22 | 0 | 7 | 29 |
| [frontend/src/api/settings.ts](/frontend/src/api/settings.ts) | TypeScript | 26 | 0 | 6 | 32 |
| [frontend/src/api/storage.ts](/frontend/src/api/storage.ts) | TypeScript | 19 | 4 | 5 | 28 |
| [frontend/src/api/system.ts](/frontend/src/api/system.ts) | TypeScript | 30 | 0 | 6 | 36 |
| [frontend/src/api/task.ts](/frontend/src/api/task.ts) | TypeScript | 35 | 7 | 9 | 51 |
| [frontend/src/api/token.ts](/frontend/src/api/token.ts) | TypeScript | 0 | 0 | 1 | 1 |
| [frontend/src/api/user.ts](/frontend/src/api/user.ts) | TypeScript | 42 | 4 | 8 | 54 |
| [frontend/src/env.d.ts](/frontend/src/env.d.ts) | TypeScript | 0 | 1 | 1 | 2 |
| [frontend/src/i18n/index.ts](/frontend/src/i18n/index.ts) | TypeScript | 14 | 0 | 4 | 18 |
| [frontend/src/i18n/locales/en.json](/frontend/src/i18n/locales/en.json) | JSON | 565 | 0 | 1 | 566 |
| [frontend/src/i18n/locales/zh-CN.json](/frontend/src/i18n/locales/zh-CN.json) | JSON | 565 | 0 | 1 | 566 |
| [frontend/src/main.ts](/frontend/src/main.ts) | TypeScript | 15 | 0 | 5 | 20 |
| [frontend/src/mock/dashboard.ts](/frontend/src/mock/dashboard.ts) | TypeScript | 23 | 0 | 5 | 28 |
| [frontend/src/mock/provider.ts](/frontend/src/mock/provider.ts) | TypeScript | 6 | 0 | 2 | 8 |
| [frontend/src/mock/quota.ts](/frontend/src/mock/quota.ts) | TypeScript | 6 | 0 | 2 | 8 |
| [frontend/src/mock/tasks.ts](/frontend/src/mock/tasks.ts) | TypeScript | 0 | 0 | 1 | 1 |
| [frontend/src/router/index.ts](/frontend/src/router/index.ts) | TypeScript | 45 | 0 | 8 | 53 |
| [frontend/src/stores/console.ts](/frontend/src/stores/console.ts) | TypeScript | 523 | 30 | 66 | 619 |
| [frontend/src/styles/index.css](/frontend/src/styles/index.css) | PostCSS | 604 | 2 | 108 | 714 |
| [frontend/src/types/api.ts](/frontend/src/types/api.ts) | TypeScript | 5 | 0 | 0 | 5 |
| [frontend/src/types/common.ts](/frontend/src/types/common.ts) | TypeScript | 8 | 0 | 2 | 10 |
| [frontend/src/types/dashboard.ts](/frontend/src/types/dashboard.ts) | TypeScript | 24 | 0 | 5 | 29 |
| [frontend/src/types/download.ts](/frontend/src/types/download.ts) | TypeScript | 32 | 2 | 3 | 37 |
| [frontend/src/types/mount.ts](/frontend/src/types/mount.ts) | TypeScript | 25 | 0 | 3 | 28 |
| [frontend/src/types/provider.ts](/frontend/src/types/provider.ts) | TypeScript | 20 | 0 | 1 | 21 |
| [frontend/src/types/quota.ts](/frontend/src/types/quota.ts) | TypeScript | 18 | 0 | 1 | 19 |
| [frontend/src/types/rclone.ts](/frontend/src/types/rclone.ts) | TypeScript | 25 | 0 | 3 | 28 |
| [frontend/src/types/task.ts](/frontend/src/types/task.ts) | TypeScript | 0 | 0 | 1 | 1 |
| [frontend/src/types/token.ts](/frontend/src/types/token.ts) | TypeScript | 0 | 0 | 1 | 1 |
| [frontend/src/utils/request.ts](/frontend/src/utils/request.ts) | TypeScript | 30 | 2 | 4 | 36 |
| [frontend/src/utils/session.ts](/frontend/src/utils/session.ts) | TypeScript | 78 | 0 | 14 | 92 |
| [frontend/tsconfig.json](/frontend/tsconfig.json) | JSON with Comments | 26 | 0 | 1 | 27 |
| [frontend/tsconfig.node.json](/frontend/tsconfig.node.json) | JSON | 9 | 0 | 1 | 10 |
| [frontend/vite.config.ts](/frontend/vite.config.ts) | TypeScript | 21 | 0 | 1 | 22 |
| [package-lock.json](/package-lock.json) | JSON | 6 | 0 | 1 | 7 |

[Summary](results.md) / Details / [Diff Summary](diff.md) / [Diff Details](diff-details.md)