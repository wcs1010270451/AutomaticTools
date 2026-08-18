package com.automatictools.app

import android.app.Activity
import android.app.AlertDialog
import android.content.Intent
import android.graphics.Color
import android.graphics.drawable.GradientDrawable
import android.net.Uri
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.provider.Settings
import android.text.InputType
import android.view.Gravity
import android.view.View
import android.widget.Button
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.ScrollView
import android.widget.TextView
import android.widget.Toast
import org.json.JSONArray

class MainActivity : Activity() {
    private lateinit var overlayStatus: TextView
    private lateinit var accessibilityStatus: TextView
    private lateinit var contentFrame: LinearLayout
    private lateinit var navBar: LinearLayout

    private var currentTab = 0 // 0: Home, 1: Tools, 2: Profile
    private val mainHandler = Handler(Looper.getMainLooper())

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(buildRootView())
        switchTab(0)
    }

    override fun onResume() {
        super.onResume()
        if (::overlayStatus.isInitialized) {
            refreshStatuses()
        }
    }

    // ─── Root layout ───────────────────────────────────────────────────

    private fun buildRootView(): View {
        val root = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setBackgroundColor(Color.rgb(239, 246, 255))
        }

        contentFrame = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, 0, 1f
            )
        }
        root.addView(contentFrame)

        navBar = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            setBackgroundColor(Color.WHITE)
            elevation = dp(8).toFloat()
            setPadding(0, dp(8), 0, dp(8))
        }
        navBar.addView(navItem("主页", 0), lp(0, 1f))
        navBar.addView(navItem("工具", 1), lp(0, 1f))
        navBar.addView(navItem("个人", 2), lp(0, 1f))
        root.addView(navBar, matchWrap())

        return root
    }

    private fun navItem(label: String, index: Int): TextView = TextView(this).apply {
        text = label
        textSize = 14f
        gravity = Gravity.CENTER
        setTextColor(if (currentTab == index) Color.rgb(37, 99, 235) else Color.GRAY)
        setOnClickListener { switchTab(index) }
    }

    private fun switchTab(index: Int) {
        currentTab = index
        contentFrame.removeAllViews()
        for (i in 0 until navBar.childCount) {
            val tv = navBar.getChildAt(i) as TextView
            tv.setTextColor(if (i == index) Color.rgb(37, 99, 235) else Color.GRAY)
        }
        when (index) {
            0 -> { contentFrame.addView(buildHomeView()); refreshStatuses() }
            1 -> buildToolsView()
            2 -> contentFrame.addView(buildProfileView())
        }
    }

    // ─── Tab 0: Home (auto-clicker) ────────────────────────────────────

    private fun buildHomeView(): View {
        val scroll = scrollable {
            setPadding(dp(24), dp(28), dp(24), dp(24))
        }
        val container = scroll.content

        container.addView(TextView(this).apply {
            text = "安卓自动点击器"
            textSize = 22f
            setTextColor(Color.rgb(15, 23, 42))
            gravity = Gravity.CENTER
        }, matchWrap())

        container.addView(TextView(this).apply {
            text = "先开启悬浮窗和无障碍权限，然后启动悬浮控制面板。"
            textSize = 15f
            setTextColor(Color.rgb(51, 65, 85))
            setPadding(0, dp(12), 0, dp(18))
        }, matchWrap())

        overlayStatus = statusText()
        accessibilityStatus = statusText()
        container.addView(overlayStatus, matchWrap())
        container.addView(accessibilityStatus, matchWrap())

        container.addView(actionButton("开启悬浮窗权限").apply {
            setOnClickListener { openOverlaySettings() }
        }, matchWrap(18))

        container.addView(actionButton("开启无障碍服务").apply {
            setOnClickListener { startActivity(Intent(Settings.ACTION_ACCESSIBILITY_SETTINGS)) }
        }, matchWrap(10))

        container.addView(actionButton("启动悬浮控制").apply {
            setOnClickListener {
                if (Settings.canDrawOverlays(this@MainActivity)) {
                    startService(Intent(this@MainActivity, OverlayService::class.java))
                } else {
                    Toast.makeText(this@MainActivity, "请先开启悬浮窗权限", Toast.LENGTH_SHORT).show()
                }
                refreshStatuses()
            }
        }, matchWrap(18))

        container.addView(TextView(this).apply {
            text = "拖动准星到目标位置，点\"锁定\"，再点\"开始\"。运行后面板会自动收起。"
            textSize = 13f
            setTextColor(Color.rgb(71, 85, 105))
            setPadding(0, dp(18), 0, 0)
        }, matchWrap())

        return scroll.root
    }

    // ─── Tab 1: Tools + License Code ───────────────────────────────────

    private fun buildToolsView() {
        val scroll = scrollable {
            setPadding(dp(24), dp(24), dp(24), dp(24))
        }
        val container = scroll.content

        container.addView(TextView(this).apply {
            text = "工具列表"
            textSize = 20f
            setTextColor(Color.rgb(15, 23, 42))
            setPadding(0, 0, 0, dp(4))
        }, matchWrap())

        val loadingLabel = TextView(this).apply {
            text = "加载中..."
            textSize = 14f
            setTextColor(Color.GRAY)
            setPadding(0, dp(12), 0, dp(12))
            gravity = Gravity.CENTER
        }
        container.addView(loadingLabel, matchWrap())

        val toolsContainer = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
        }
        container.addView(toolsContainer, matchWrap())

        // License code section
        container.addView(cardDivider("授权码兑换"), matchWrap(24))

        val codeInput = EditText(this).apply {
            hint = "输入授权码 AT-XXXX-XXXX-XXXX-XXXX"
            textSize = 14f
            setPadding(dp(12), dp(12), dp(12), dp(12))
            background = GradientDrawable().apply {
                setColor(Color.WHITE)
                cornerRadius = dp(10).toFloat()
                setStroke(dp(1).toInt(), Color.rgb(203, 213, 225))
            }
        }
        container.addView(codeInput, matchWrap(8))

        container.addView(actionButton("兑换授权码").apply {
            setOnClickListener {
                val code = codeInput.text.toString().trim()
                if (code.isBlank()) {
                    Toast.makeText(this@MainActivity, "请输入授权码", Toast.LENGTH_SHORT).show()
                    return@setOnClickListener
                }
                if (!UserManager.isLoggedIn(this@MainActivity)) {
                    Toast.makeText(this@MainActivity, "请先登录", Toast.LENGTH_SHORT).show()
                    return@setOnClickListener
                }
                it.isEnabled = false
                UserManager.redeemCode(this@MainActivity, code) { success, msg, _ ->
                    it.isEnabled = true
                    Toast.makeText(this@MainActivity, msg, Toast.LENGTH_SHORT).show()
                    if (success) {
                        codeInput.setText("")
                        // Refresh tools list
                        buildToolsView()
                    }
                }
            }
        }, matchWrap(10))

        container.addView(TextView(this).apply {
            text = "授权码可在淘宝或咸鱼购买，兑换后即可解锁对应工具。"
            textSize = 12f
            setTextColor(Color.rgb(148, 163, 184))
            setPadding(0, dp(8), 0, 0)
        }, matchWrap())

        contentFrame.addView(scroll.root)

        // Fetch tools in background
        Thread {
            try {
                val toolsResult = ApiClient.listTools()
                val tools = toolsResult.optJSONArray("tools") ?: JSONArray()

                // Fetch entitlements if logged in
                var entitledCodes = setOf<String>()
                val token = UserManager.getToken(this)
                if (token != null) {
                    try {
                        val entResult = ApiClient.myEntitlements(token)
                        val ents = entResult.optJSONArray("entitlements") ?: JSONArray()
                        val codes = mutableSetOf<String>()
                        for (i in 0 until ents.length()) {
                            codes.add(ents.getJSONObject(i).optString("toolCode"))
                        }
                        entitledCodes = codes
                    } catch (_: Exception) { /* ignore */ }
                }

                mainHandler.post {
                    loadingLabel.visibility = View.GONE
                    toolsContainer.removeAllViews()

                    if (tools.length() == 0) {
                        toolsContainer.addView(TextView(this).apply {
                            text = "暂无可用工具"
                            textSize = 14f
                            setTextColor(Color.GRAY)
                            gravity = Gravity.CENTER
                            setPadding(0, dp(20), 0, dp(20))
                        }, matchWrap())
                        return@post
                    }

                    for (i in 0 until tools.length()) {
                        val tool = tools.getJSONObject(i)
                        val code = tool.optString("code")
                        val name = tool.optString("name")
                        val desc = tool.optString("description")
                        val owned = entitledCodes.contains(code)
                        toolsContainer.addView(toolCard(name, desc, owned), matchWrap(8))
                    }
                }
            } catch (e: Exception) {
                mainHandler.post {
                    loadingLabel.text = "加载失败，请下拉刷新"
                    loadingLabel.setTextColor(Color.rgb(220, 38, 38))
                }
            }
        }.start()
    }

    private fun toolCard(name: String, desc: String, owned: Boolean): View {
        val card = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(dp(16), dp(14), dp(16), dp(14))
            background = GradientDrawable().apply {
                setColor(Color.WHITE)
                cornerRadius = dp(12).toFloat()
            }
            elevation = dp(2).toFloat()
        }

        val headerRow = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
        }
        headerRow.addView(TextView(this).apply {
            text = name
            textSize = 16f
            setTextColor(Color.rgb(15, 23, 42))
        }, LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f))

        headerRow.addView(TextView(this).apply {
            text = if (owned) "已解锁" else "未解锁"
            textSize = 12f
            setTextColor(if (owned) Color.rgb(22, 163, 74) else Color.rgb(148, 163, 184))
            setPadding(dp(8), dp(2), dp(8), dp(2))
            background = GradientDrawable().apply {
                cornerRadius = dp(6).toFloat()
                setColor(if (owned) Color.rgb(220, 252, 231) else Color.rgb(241, 245, 249))
            }
        })
        card.addView(headerRow)

        if (desc.isNotBlank()) {
            card.addView(TextView(this).apply {
                text = desc
                textSize = 13f
                setTextColor(Color.rgb(100, 116, 139))
                setPadding(0, dp(6), 0, 0)
            }, matchWrap())
        }
        return card
    }

    // ─── Tab 2: Profile / Login / Register ─────────────────────────────

    private fun buildProfileView(): View {
        val scroll = scrollable {
            orientation = LinearLayout.VERTICAL
            setPadding(dp(24), dp(32), dp(24), dp(24))
        }
        val container = scroll.content

        if (UserManager.isLoggedIn(this)) {
            buildLoggedInView(container)
        } else {
            buildLoginView(container)
        }
        return scroll.root
    }

    private fun buildLoginView(container: LinearLayout) {
        // Title
        container.addView(TextView(this).apply {
            text = "登录账号"
            textSize = 22f
            setTextColor(Color.rgb(15, 23, 42))
            gravity = Gravity.CENTER
            setPadding(0, dp(20), 0, dp(4))
        }, matchWrap())

        container.addView(TextView(this).apply {
            text = "登录后同步配置、兑换授权码"
            textSize = 14f
            setTextColor(Color.rgb(100, 116, 139))
            gravity = Gravity.CENTER
            setPadding(0, 0, 0, dp(28))
        }, matchWrap())

        // Login form
        val accountEdit = EditText(this).apply {
            hint = "邮箱 / 用户名 / 手机号"
            textSize = 14f
            setPadding(dp(12), dp(12), dp(12), dp(12))
            background = whiteInput()
        }
        container.addView(accountEdit, matchWrap(0))

        val passEdit = EditText(this).apply {
            hint = "密码"
            inputType = InputType.TYPE_CLASS_TEXT or InputType.TYPE_TEXT_VARIATION_PASSWORD
            textSize = 14f
            setPadding(dp(12), dp(12), dp(12), dp(12))
            background = whiteInput()
        }
        container.addView(passEdit, matchWrap(10))

        val loginBtn = actionButton("登  录")
        container.addView(loginBtn, matchWrap(16))

        loginBtn.setOnClickListener {
            val account = accountEdit.text.toString().trim()
            val password = passEdit.text.toString()
            if (account.isBlank() || password.isBlank()) {
                Toast.makeText(this, "请填写账号和密码", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            loginBtn.isEnabled = false
            loginBtn.text = "登录中..."
            UserManager.login(this, account, password) { success, msg ->
                loginBtn.isEnabled = true
                loginBtn.text = "登  录"
                Toast.makeText(this, msg, Toast.LENGTH_SHORT).show()
                if (success) switchTab(2)
            }
        }

        // Divider
        container.addView(TextView(this).apply {
            text = "── 还没有账号？ ──"
            textSize = 13f
            setTextColor(Color.rgb(148, 163, 184))
            gravity = Gravity.CENTER
            setPadding(0, dp(24), 0, dp(16))
        }, matchWrap())

        // Register button
        container.addView(actionButton("注册新账号").apply {
            background = GradientDrawable().apply {
                cornerRadius = dp(12).toFloat()
                setColor(Color.rgb(241, 245, 249))
            }
            setTextColor(Color.rgb(37, 99, 235))
        }.apply {
            setOnClickListener { showRegisterDialog() }
        }, matchWrap())
    }

    private fun buildLoggedInView(container: LinearLayout) {
        // Avatar circle
        container.addView(View(this).apply {
            layoutParams = LinearLayout.LayoutParams(dp(72), dp(72)).apply {
                gravity = Gravity.CENTER_HORIZONTAL
                bottomMargin = dp(12)
            }
            background = GradientDrawable().apply {
                shape = GradientDrawable.OVAL
                setColor(Color.rgb(37, 99, 235))
            }
        })

        val username = UserManager.getUsername(this) ?: ""
        val email = UserManager.getEmail(this) ?: ""

        container.addView(TextView(this).apply {
            text = username.ifBlank { "用户" }
            textSize = 20f
            setTextColor(Color.rgb(15, 23, 42))
            gravity = Gravity.CENTER
            setPadding(0, 0, 0, dp(4))
        }, matchWrap())

        container.addView(TextView(this).apply {
            text = email
            textSize = 14f
            setTextColor(Color.rgb(100, 116, 139))
            gravity = Gravity.CENTER
            setPadding(0, 0, 0, dp(28))
        }, matchWrap())

        // Menu items
        container.addView(profileMenuItem("同步云端数据", "从服务器拉取最新账号信息") {
            UserManager.fetchMe(this) { success, msg ->
                Toast.makeText(this, msg, Toast.LENGTH_SHORT).show()
                if (success) switchTab(2)
            }
        })

        container.addView(profileMenuItem("兑换授权码", "输入授权码解锁工具") {
            switchTab(1)
        })

        container.addView(profileMenuItem("使用教程", "查看使用帮助") {
            Toast.makeText(this, "教程功能开发中", Toast.LENGTH_SHORT).show()
        })

        container.addView(profileMenuItem("关于我们", "版本信息与联系方式") {
            Toast.makeText(this, "AutomaticTools v1.0", Toast.LENGTH_SHORT).show()
        })

        // Logout
        container.addView(actionButton("退出登录").apply {
            background = GradientDrawable().apply {
                cornerRadius = dp(12).toFloat()
                setColor(Color.rgb(220, 38, 38))
            }
        }.apply {
            setOnClickListener {
                AlertDialog.Builder(this@MainActivity)
                    .setTitle("退出登录")
                    .setMessage("确定要退出登录吗？")
                    .setPositiveButton("退出") { _, _ ->
                        UserManager.logout(this@MainActivity)
                        switchTab(2)
                    }
                    .setNegativeButton("取消", null)
                    .show()
            }
        }, matchWrap(32))
    }

    private fun profileMenuItem(title: String, subtitle: String, onClick: () -> Unit): View {
        val row = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(dp(16), dp(14), dp(16), dp(14))
            background = GradientDrawable().apply {
                setColor(Color.WHITE)
                cornerRadius = dp(10).toFloat()
            }
            layoutParams = matchWrap(8)
            setOnClickListener { onClick() }
        }

        val textCol = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
        }
        textCol.addView(TextView(this).apply {
            text = title
            textSize = 15f
            setTextColor(Color.rgb(30, 41, 59))
        })
        textCol.addView(TextView(this).apply {
            text = subtitle
            textSize = 12f
            setTextColor(Color.rgb(148, 163, 184))
            setPadding(0, dp(2), 0, 0)
        })
        row.addView(textCol, LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f))

        row.addView(TextView(this).apply {
            text = ">"
            textSize = 16f
            setTextColor(Color.rgb(203, 213, 225))
        })
        return row
    }

    // ─── Register Dialog ───────────────────────────────────────────────

    private fun showRegisterDialog() {
        val layout = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(dp(20), dp(16), dp(20), dp(16))
        }

        val emailEdit = EditText(this).apply { hint = "邮箱地址"; inputType = InputType.TYPE_CLASS_TEXT or InputType.TYPE_TEXT_VARIATION_EMAIL_ADDRESS }
        val codeEdit = EditText(this).apply { hint = "邮箱验证码" }
        val passEdit = EditText(this).apply { hint = "设置密码（至少6位）"; inputType = InputType.TYPE_CLASS_TEXT or InputType.TYPE_TEXT_VARIATION_PASSWORD }
        val nameEdit = EditText(this).apply { hint = "用户名（选填）" }

        val sendCodeBtn = Button(this).apply {
            text = "发送验证码"
            textSize = 13f
            setTextColor(Color.WHITE)
            background = GradientDrawable().apply {
                cornerRadius = dp(8).toFloat()
                setColor(Color.rgb(100, 116, 139))
            }
        }

        val codeRow = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
        }
        codeRow.addView(codeEdit, LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f).apply {
            rightMargin = dp(8)
        })
        codeRow.addView(sendCodeBtn, LinearLayout.LayoutParams(LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT))

        layout.addView(emailEdit)
        layout.addView(codeRow)
        layout.addView(passEdit)
        layout.addView(nameEdit)

        var codeSent = false

        sendCodeBtn.setOnClickListener {
            val email = emailEdit.text.toString().trim()
            if (email.isBlank() || !email.contains("@")) {
                Toast.makeText(this, "请输入有效邮箱", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            sendCodeBtn.isEnabled = false
            sendCodeBtn.text = "发送中..."
            UserManager.sendEmailCode(email) { success, msg ->
                sendCodeBtn.isEnabled = true
                sendCodeBtn.text = "发送验证码"
                Toast.makeText(this, msg, Toast.LENGTH_SHORT).show()
                if (success) {
                    codeSent = true
                    sendCodeBtn.isEnabled = false
                    sendCodeBtn.text = "已发送"
                }
            }
        }

        AlertDialog.Builder(this)
            .setTitle("注册新账号")
            .setView(layout)
            .setPositiveButton("注  册") { _, _ ->
                val email = emailEdit.text.toString().trim()
                val code = codeEdit.text.toString().trim()
                val pass = passEdit.text.toString()
                val name = nameEdit.text.toString().trim()

                if (email.isBlank() || code.isBlank() || pass.isBlank()) {
                    Toast.makeText(this, "请填写必要信息", Toast.LENGTH_SHORT).show()
                    return@setPositiveButton
                }
                if (pass.length < 6) {
                    Toast.makeText(this, "密码至少6位", Toast.LENGTH_SHORT).show()
                    return@setPositiveButton
                }
                UserManager.register(this, email, code, pass, name) { success, msg ->
                    Toast.makeText(this, msg, Toast.LENGTH_SHORT).show()
                    if (success) switchTab(2)
                }
            }
            .setNegativeButton("取消", null)
            .show()
    }

    // ─── Helpers ────────────────────────────────────────────────────────

    private fun refreshStatuses() {
        overlayStatus.text = if (Settings.canDrawOverlays(this)) "悬浮窗权限：已开启" else "悬浮窗权限：未开启"
        accessibilityStatus.text = if (ClickAccessibilityService.isRunning) "无障碍服务：已开启" else "无障碍服务：未开启"
    }

    private fun openOverlaySettings() {
        startActivity(Intent(Settings.ACTION_MANAGE_OVERLAY_PERMISSION, Uri.parse("package:$packageName")))
    }

    private fun cardDivider(title: String): View {
        val row = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(0, dp(4), 0, dp(4))
        }
        row.addView(View(this).apply {
            layoutParams = LinearLayout.LayoutParams(0, dp(1), 0f)
            setBackgroundColor(Color.rgb(226, 232, 240))
        })
        row.addView(TextView(this).apply {
            text = "  $title  "
            textSize = 13f
            setTextColor(Color.rgb(100, 116, 139))
        })
        row.addView(View(this).apply {
            layoutParams = LinearLayout.LayoutParams(0, dp(1), 0f)
            setBackgroundColor(Color.rgb(226, 232, 240))
        })
        return row
    }

    private fun statusText(): TextView = TextView(this).apply {
        textSize = 15f
        setTextColor(Color.rgb(30, 41, 59))
        setPadding(0, dp(4), 0, dp(4))
    }

    private fun actionButton(textValue: String): Button = Button(this).apply {
        text = textValue
        textSize = 15f
        setTextColor(Color.WHITE)
        background = GradientDrawable().apply {
            cornerRadius = dp(12).toFloat()
            setColor(Color.rgb(37, 99, 235))
        }
    }

    private fun whiteInput(): GradientDrawable = GradientDrawable().apply {
        setColor(Color.WHITE)
        cornerRadius = dp(10).toFloat()
        setStroke(dp(1).toInt(), Color.rgb(203, 213, 225))
    }

    private data class ScrollableContent(
        val root: ScrollView,
        val content: LinearLayout,
    )

    private fun scrollable(init: LinearLayout.() -> Unit): ScrollableContent {
        val inner = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            init()
        }
        val root = ScrollView(this).apply {
            addView(inner, LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                LinearLayout.LayoutParams.WRAP_CONTENT
            ))
        }
        return ScrollableContent(root, inner)
    }

    private fun matchWrap(top: Int = 0): LinearLayout.LayoutParams =
        LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.MATCH_PARENT,
            LinearLayout.LayoutParams.WRAP_CONTENT
        ).apply { topMargin = dp(top) }

    private fun lp(height: Int, weight: Float): LinearLayout.LayoutParams =
        LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, weight)

    private fun dp(value: Int): Int =
        (value * resources.displayMetrics.density).toInt()
}
