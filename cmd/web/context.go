package main

type contextKey string

const isAuthenticatedContextKey = contextKey("isAuthenticated")

//因為存在應用程式使用的其他第三方套件也希望使用“isAuthenticated”鍵儲存資料的風險 - 這會導致命名衝突。
// 為了避免這種情況，最好建立自己的自訂類型，並將其用作上下文鍵。
