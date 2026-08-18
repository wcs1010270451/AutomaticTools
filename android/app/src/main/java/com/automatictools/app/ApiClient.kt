package com.automatictools.app

import android.os.Build
import org.json.JSONArray
import org.json.JSONObject
import java.io.BufferedReader
import java.io.InputStreamReader
import java.io.OutputStreamWriter
import java.net.HttpURLConnection
import java.net.URL
import java.nio.charset.StandardCharsets

class ApiError(message: String, val status: Int = 0) : Exception(message)
class NetworkError(message: String) : Exception(message)

object ApiClient {
    private const val BASE_URL = "https://autumnwind.top"
    private const val TIMEOUT_MS = 15_000

    fun sendEmailCode(email: String): JSONObject {
        val body = JSONObject().put("email", email)
        return post("/api/auth/email-code", body)
    }

    fun register(
        email: String,
        emailCode: String,
        password: String,
        username: String,
        deviceId: String,
    ): JSONObject {
        val body = JSONObject().apply {
            put("email", email)
            put("emailCode", emailCode)
            put("password", password)
            put("deviceId", deviceId)
            put("deviceName", Build.MODEL ?: "Android")
            put("platform", "android")
            if (username.isNotBlank()) put("username", username)
        }
        return post("/api/auth/register", body)
    }

    fun login(account: String, password: String, deviceId: String): JSONObject {
        val body = JSONObject().apply {
            put("account", account)
            put("password", password)
            put("deviceId", deviceId)
            put("deviceName", Build.MODEL ?: "Android")
            put("platform", "android")
        }
        return post("/api/auth/login", body)
    }

    fun me(token: String): JSONObject {
        return get("/api/me", token)
    }

    fun listTools(): JSONObject {
        return get("/api/tools")
    }

    fun myEntitlements(token: String): JSONObject {
        return get("/api/me/entitlements", token)
    }

    fun myPurchases(token: String): JSONObject {
        return get("/api/me/purchases", token)
    }

    fun redeemLicenseCode(token: String, code: String): JSONObject {
        val body = JSONObject().put("code", code)
        return post("/api/license-codes/redeem", body, token)
    }

    // --- HTTP primitives ---

    private fun get(path: String, token: String = ""): JSONObject {
        val url = URL("$BASE_URL$path")
        val conn = (url.openConnection() as HttpURLConnection).apply {
            requestMethod = "GET"
            connectTimeout = TIMEOUT_MS
            readTimeout = TIMEOUT_MS
            setRequestProperty("Accept", "application/json")
            setRequestProperty("User-Agent", "AutomaticTools-Android/1.0")
            if (token.isNotBlank()) setRequestProperty("Authorization", "Bearer $token")
        }
        return readResponse(conn)
    }

    private fun post(path: String, body: JSONObject, token: String = ""): JSONObject {
        val url = URL("$BASE_URL$path")
        val conn = (url.openConnection() as HttpURLConnection).apply {
            requestMethod = "POST"
            connectTimeout = TIMEOUT_MS
            readTimeout = TIMEOUT_MS
            doOutput = true
            setRequestProperty("Accept", "application/json")
            setRequestProperty("Content-Type", "application/json; charset=utf-8")
            setRequestProperty("User-Agent", "AutomaticTools-Android/1.0")
            if (token.isNotBlank()) setRequestProperty("Authorization", "Bearer $token")
        }
        val payload = body.toString().toByteArray(StandardCharsets.UTF_8)
        OutputStreamWriter(conn.outputStream, StandardCharsets.UTF_8).use { it.write(String(payload, StandardCharsets.UTF_8)) }
        return readResponse(conn)
    }

    private fun readResponse(conn: HttpURLConnection): JSONObject {
        val stream = try {
            conn.inputStream
        } catch (e: Exception) {
            conn.errorStream
        }
        val text = BufferedReader(InputStreamReader(stream, StandardCharsets.UTF_8)).use { it.readText() }
        val code = conn.responseCode
        conn.disconnect()

        if (text.isBlank()) {
            if (code in 200..299) return JSONObject()
            throw ApiError("服务器请求失败，请稍后重试。", code)
        }

        val json = try {
            JSONObject(text)
        } catch (_: Exception) {
            if (code in 200..299) return JSONObject()
            throw ApiError("服务器返回了无法识别的数据。", code)
        }

        if (code !in 200..299) {
            val message = json.optString("error", "").takeIf { it.isNotBlank() } ?: "服务器请求失败，请稍后重试。"
            throw ApiError(message, code)
        }
        return json
    }
}
