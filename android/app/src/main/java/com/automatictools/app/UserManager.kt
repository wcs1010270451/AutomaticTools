package com.automatictools.app

import android.content.Context
import android.content.SharedPreferences

object UserManager {
    private const val PREFS_NAME = "user_prefs"
    private const val KEY_IS_LOGGED_IN = "is_logged_in"
    private const val KEY_USERNAME = "username"
    private const val KEY_PASSWORD = "password" // For fake demo only!

    private fun getPrefs(context: Context): SharedPreferences {
        return context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
    }

    fun isLoggedIn(context: Context): Boolean {
        return getPrefs(context).getBoolean(KEY_IS_LOGGED_IN, false)
    }

    fun getUsername(context: Context): String? {
        return getPrefs(context).getString(KEY_USERNAME, null)
    }

    // Fake Register
    fun register(context: Context, user: String, pass: String): Boolean {
        if (user.isEmpty() || pass.isEmpty()) return false
        getPrefs(context).edit().apply {
            putString(KEY_USERNAME, user)
            putString(KEY_PASSWORD, pass)
            apply()
        }
        return true
    }

    // Fake Login
    fun login(context: Context, user: String, pass: String, callback: (Boolean, String) -> Unit) {
        val prefs = getPrefs(context)
        val savedUser = prefs.getString(KEY_USERNAME, null)
        val savedPass = prefs.getString(KEY_PASSWORD, null)

        // Simulate network delay
        android.os.Handler(android.os.Looper.getMainLooper()).postDelayed({
            if (savedUser == user && savedPass == pass) {
                prefs.edit().putBoolean(KEY_IS_LOGGED_IN, true).apply()
                callback(true, "登录成功")
            } else {
                callback(false, "用户名或密码错误")
            }
        }, 1000)
    }

    fun logout(context: Context) {
        getPrefs(context).edit().putBoolean(KEY_IS_LOGGED_IN, false).apply()
    }
}
