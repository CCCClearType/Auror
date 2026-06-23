import os
with open('api/api_code_mapping.md', 'a', encoding='utf-8') as f:
    f.write("\n---\n\n## 6. 系統狀態與監測 (System Status & iLearn Monitoring)\n\n")
    f.write("| HTTP 方法 | API 網址路徑 | 路由註冊 (Router) | 對應的控制器函式 (Controller) | 備註功能 |\n")
    f.write("|---|---|---|---|---|\n")
    f.write("| **GET** | `/api/ilearn-status` | `api.GET(\"/ilearn-status\", ...)` | [status_controller.go:13](file:///c:/Users/HP/Downloads/dbms-git/dbms/backend/controllers/status_controller.go#L13) | 即時戳 iLearn 確認目前狀態與延遲 |\n")
    f.write("| **POST** | `/api/ilearn-reports` | `api.POST(\"/ilearn-reports\", ...)` | [status_controller.go:40](file:///c:/Users/HP/Downloads/dbms-git/dbms/backend/controllers/status_controller.go#L40) | 使用者回報 iLearn 異常 |\n")
    f.write("| **GET** | `/api/ilearn-history` | `api.GET(\"/ilearn-history\", ...)` | [status_controller.go:65](file:///c:/Users/HP/Downloads/dbms-git/dbms/backend/controllers/status_controller.go#L65) | 取得過去一段時間的連線紀錄與統計 |\n")
