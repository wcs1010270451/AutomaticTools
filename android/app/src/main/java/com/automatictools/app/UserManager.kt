package com.automatictools.app

import android.content.Context
import android.content.SharedPreferences
import android.os.Handler
import android.os.Looper
import android.provider.Settings
import org.json.JSONObject
import java.util.UUID

object UserManager {
    private const val PREFS_NAME = "user_prefs"
    private const val KEY_TOKEN = "auth_token"
    private const val KEY_USER_ID = "user_id"
    private const val KEY_USERNAME = "username"
    private const val KEY_EMAIL = "email"
    private const val KEY_DEVICE_ID = "device_id"

    private fun prefs(context: Context): SharedPreferences =
        context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    fun isLoggedIn(context: Context): Boolean =
        prefs(context).getString(KEY_TOKEN, null)?.isNotBlank() == true

    fun getToken(context: Context): String? =
        prefs(context).getString(KEY_TOKEN, null)

    fun getUsername(context: Context): String? =
        prefs(context).getString(KEY_USERNAME, null)

    fun getEmail(context: Context): String? =
        prefs(context).getString(KEY_EMAIL, null)

    fun getUserId(context: Context): Long =
        prefs(context).getLong(KEY_USER_ID, 0L)

    fun getDeviceId(context: Context): String {
        val p = prefs(context)
        val existing = p.getString(KEY_DEVICE_ID, null)
        if (existing != null) return existing
        val id = UUID.randomUUID().toString()
        p.edit().putString(KEY_DEVICE_ID, id).apply()
        return id
    }

    fun saveAuth(context: Context, token: String, user: JSONObject) {
        prefs(context).edit().apply {
            putString(KEY_TOKEN, token)
            putLong(KEY_USER_ID, user.optLong("id", 0))
            putString(KEY_USERNAME, user.optString("username", ""))
            putString(KEY_EMAIL, user.optString("email", ""))
            apply()
        }
    }

    fun updateUserInfo(context: Context, user: JSONObject) {
        prefs(context).edit().apply {
            putLong(KEY_USER_ID, user.optLong("id", 0))
            putString(KEY_USERNAME, user.optString("username", ""))
            putString(KEY_EMAIL, user.optString("email", ""))
            apply()
        }
    }

    fun logout(context: Context) {
        prefs(context).edit().apply {
            remove(KEY_TOKEN)
            remove(KEY_USER_ID)
            remove(KEY_USERNAME)
            remove(KEY_EMAIL)
            apply()
        }
    }

    // --- Async API operations (run on background thread, callback on main) ---

    private val mainHandler = Handler(Looper.getMainLooper())

    fun login(context: Context, account: String, password: String, callback: (Boolean, String) -> Unit) {
        Thread {
            try {
                val deviceId = getDeviceId(context)
                val result = ApiClient.login(account, password, deviceId)
                val token = result.getString("token")
                val user = result.getJSONObject("user")
                saveAuth(context, token, user)
                mainHandler.post { callback(true, "登录成功") }
            } catch (e: ApiError) {
                mainHandler.post { callback(false, e.message ?: "登录失败") }
            } catch (e: Exception) {
                mainHandler.post { callback(false, "无法连接服务器，请检查网络。") }
            }
        }.start()
    }

    fun register(
        context: Context,
        email: String,
        emailCode: String,
        password: String,
        username: String,
        callback: (Boolean, String) -> Unit,
    ) {
        Thread {
            try {
                val deviceId = getDeviceId(context)
                val result = ApiClient.register(email, emailCode, password, username, deviceId)
                val token = result.getString("token")
                val user = result.getJSONObject("user")
                saveAuth(context, token, user)
                mainHandler.post { callback(true, "注册成功") }
            } catch (e: ApiError) {
                mainHandler.post { callback(false, e.message ?: "注册失败") }
            } catch (e: Exception) {
                mainHandler.post { callback(false, "无法连接服务器，请检查网络。") }
            }
        }.start()
    }

    fun sendEmailCode(email: String, callback: (Boolean, String) -> Unit) {
        Thread {
            try {
                ApiClient.sendEmailCode(email)
                mainHandler.post { callback(true, "验证码已发送") }
            } catch (e: ApiError) {
                mainHandler.post { callback(false, e.message ?: "发送失败") }
            } catch (e: Exception) {
                mainHandler.post { callback(false, "无法连接服务器，请检查网络。") }
            }
        }.start()
    }

    fun fetchMe(context: Context, callback: (Boolean, String) -> Unit) {
        val token = getToken(context) ?: return
        Thread {
            try {
                val result = ApiClient.me(token)
                updateUserInfo(context, result)
                mainHandler.post { callback(true, "同步成功") }
            } catch (e: ApiError) {
                mainHandler.post { callback(false, e.message ?: "同步失败") }
            } catch (e: Exception) {
                mainHandler.post { callback(false, "无法连接服务器，请检查网络。") }
            }
        }.start()
    }

    fun redeemCode(context: Context, code: String, callback: (Boolean, String, JSONObject?) -> Unit) {
        val token = getToken(context) ?: run {
            callback(false, "请先登录", null)
            return
        }
        Thread {
            try {
                val result = ApiClient.redeemLicenseCode(token, code)
                mainHandler.post { callback(true, "兑换成功", result) }
            } catch (e: ApiError) {
                mainHandler.post { callback(false, e.message ?: "兑换失败", null) }
            } catch (e: Exception) {
                mainHandler.post { callback(false, "无法连接服务器，请检查网络。", null) }
            }
        }.start()
    }
}
