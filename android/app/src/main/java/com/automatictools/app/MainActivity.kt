package com.automatictools.app

import android.app.Activity
import android.app.AlertDialog
import android.content.Intent
import android.graphics.Color
import android.graphics.drawable.GradientDrawable
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.provider.Settings
import android.text.InputType
import android.view.Gravity
import android.view.View
import android.view.WindowInsets
import android.widget.Button
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.ScrollView
import android.widget.TextView
import android.widget.Toast
import org.json.JSONArray
import org.json.JSONObject

class MainActivity : Activity() {
    private lateinit var overlayStatus: TextView
    private lateinit var accessibilityStatus: TextView
    private lateinit var contentFrame: LinearLayout
    private lateinit var navBar: LinearLayout

    private var currentTab = 0 // 0: Home, 1: Tools, 2: Profile
    private var entitledToolCodes = emptySet<String>()
    private var activeToolDetail: JSONObject? = null
    private var activeToolOwned = false
    private val mainHandler = Handler(Looper.getMainLooper())

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(buildRootView())
        switchTab(0)
    }

    override fun onResume() {
        super.onResume()
        val detail = activeToolDetail
        if (currentTab == 1 && detail != null) {
            showToolDetail(detail, activeToolOwned)
        }
    }

    // ─── Root layout ───────────────────────────────────────────────────

    private fun buildRootView(): View {
        val root = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setBackgroundColor(Color.rgb(239, 246, 255))
            setOnApplyWindowInsetsListener { view, insets ->
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
                    val systemBars = insets.getInsets(WindowInsets.Type.systemBars())
                    view.setPadding(0, systemBars.top, 0, systemBars.bottom)
                } else {
                    @Suppress("DEPRECATION")
                    view.setPadding(0, insets.systemWindowInsetTop, 0, insets.systemWindowInsetBottom)
                }
                insets
            }
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
        activeToolDetail = null
        contentFrame.removeAllViews()
        for (i in 0 until navBar.childCount) {
            val tv = navBar.getChildAt(i) as TextView
            tv.setTextColor(if (i == index) Color.rgb(37, 99, 235) else Color.GRAY)
        }
        when (index) {
            0 -> contentFrame.addView(buildHomeView())
            1 -> buildToolsView()
            2 -> contentFrame.addView(buildProfileView())
        }
    }

    // ─── Tab 0: Home ───────────────────────────────────────────────────

    private fun buildHomeView(): View {
        val scroll = scrollable {
            setPadding(dp(20), dp(24), dp(20), dp(24))
        }
        val container = scroll.content
        val loggedIn = UserManager.isLoggedIn(this)
        val username = UserManager.getUsername(this).orEmpty()
        val email = UserManager.getEmail(this).orEmpty()
        if (!loggedIn) entitledToolCodes = emptySet()

        container.addView(TextView(this).apply {
            text = if (loggedIn) "你好，${username.ifBlank { "用户" }}" else "欢迎使用 AutomaticTools"
            textSize = 22f
            setTextColor(Color.rgb(15, 23, 42))
        }, matchWrap())

        container.addView(TextView(this).apply {
            text = "这里集中展示你的账户和工具开通状态。"
            textSize = 14f
            setTextColor(Color.rgb(100, 116, 139))
            setPadding(0, dp(6), 0, dp(18))
        }, matchWrap())

        val accountCard = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(dp(18), dp(16), dp(18), dp(16))
            background = cardBackground()
            setOnClickListener { switchTab(2) }
        }
        accountCard.addView(TextView(this).apply {
            text = if (loggedIn) username.ifBlank { "已登录用户" } else "尚未登录"
            textSize = 17f
            setTextColor(Color.rgb(15, 23, 42))
        }, matchWrap())
        accountCard.addView(TextView(this).apply {
            text = if (loggedIn) email.ifBlank { "账户资料已保存在本机" } else "登录后可兑换授权码并同步工具权限"
            textSize = 13f
            setTextColor(Color.rgb(100, 116, 139))
            setPadding(0, dp(5), 0, 0)
        }, matchWrap())
        accountCard.addView(TextView(this).apply {
            text = "查看账户信息  >"
            textSize = 13f
            setTextColor(Color.rgb(37, 99, 235))
            setPadding(0, dp(12), 0, 0)
        }, matchWrap())
        container.addView(accountCard, matchWrap())

        val accountStatus = TextView(this)
        val toolCount = TextView(this)
        val openMethod = TextView(this)
        val metrics = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            setPadding(0, dp(14), 0, 0)
        }
        metrics.addView(homeMetric("账户状态", accountStatus, if (loggedIn) "已登录" else "未登录"), lp(0, 1f).apply {
            rightMargin = dp(5)
        })
        metrics.addView(homeMetric("已开通工具", toolCount, "${entitledToolCodes.size} 项"), lp(0, 1f).apply {
            leftMargin = dp(5)
            rightMargin = dp(5)
        })
        metrics.addView(homeMetric("开通方式", openMethod, "授权码"), lp(0, 1f).apply {
            leftMargin = dp(5)
        })
        container.addView(metrics, matchWrap())

        val syncStatus = TextView(this).apply {
            text = if (loggedIn) "正在同步账户数据..." else "登录后显示最新工具权限"
            textSize = 12f
            setTextColor(Color.rgb(100, 116, 139))
            setPadding(0, dp(14), 0, 0)
        }
        container.addView(syncStatus, matchWrap())

        if (loggedIn) {
            val token = UserManager.getToken(this) ?: return scroll.root
            Thread {
                try {
                    val userResult = ApiClient.me(token)
                    userResult.optJSONObject("user")?.let { UserManager.updateUserInfo(this, it) }
                    val entitlementResult = ApiClient.myEntitlements(token)
                    val codes = entitlementCodes(entitlementResult.optJSONArray("entitlements") ?: JSONArray())
                    entitledToolCodes = codes
                    mainHandler.post {
                        accountStatus.text = "正常"
                        toolCount.text = "${codes.size} 项"
                        syncStatus.text = "账户数据已同步"
                        syncStatus.setTextColor(Color.rgb(22, 163, 74))
                    }
                } catch (e: ApiError) {
                    mainHandler.post {
                        if (e.status == 401) {
                            UserManager.logout(this)
                            if (currentTab == 0) switchTab(0)
                        } else {
                            syncStatus.text = e.message ?: "同步失败"
                            syncStatus.setTextColor(Color.rgb(220, 38, 38))
                        }
                    }
                } catch (_: Exception) {
                    mainHandler.post {
                        syncStatus.text = "暂时无法连接服务器"
                        syncStatus.setTextColor(Color.rgb(220, 38, 38))
                    }
                }
            }.start()
        }

        return scroll.root
    }

    private fun homeMetric(title: String, valueView: TextView, value: String): View {
        return LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(dp(12), dp(12), dp(10), dp(12))
            background = cardBackground()
            addView(TextView(this@MainActivity).apply {
                text = title
                textSize = 11f
                setTextColor(Color.rgb(100, 116, 139))
            })
            addView(valueView.apply {
                text = value
                textSize = 15f
                setTextColor(Color.rgb(30, 64, 175))
                setPadding(0, dp(6), 0, 0)
            })
        }
    }

    // ─── Tab 1: Tools ──────────────────────────────────────────────────

    private fun buildToolsView() {
        val scroll = scrollable {
            setPadding(dp(20), dp(24), dp(20), dp(24))
        }
        val container = scroll.content

        container.addView(TextView(this).apply {
            text = "工具中心"
            textSize = 22f
            setTextColor(Color.rgb(15, 23, 42))
        }, matchWrap())

        container.addView(TextView(this).apply {
            text = "选择工具查看使用方法和开通状态。"
            textSize = 14f
            setTextColor(Color.rgb(100, 116, 139))
            setPadding(0, dp(6), 0, dp(16))
        }, matchWrap())

        val loadingLabel = TextView(this).apply {
            text = "正在加载工具..."
            textSize = 14f
            setTextColor(Color.GRAY)
            setPadding(0, dp(18), 0, dp(18))
            gravity = Gravity.CENTER
        }
        container.addView(loadingLabel, matchWrap())

        val toolsContainer = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
        }
        container.addView(toolsContainer, matchWrap())
        contentFrame.addView(scroll.root)

        Thread {
            try {
                val tools = ApiClient.listTools().optJSONArray("tools") ?: JSONArray()
                val token = UserManager.getToken(this)
                entitledToolCodes = if (token == null) {
                    emptySet()
                } else {
                    entitlementCodes(ApiClient.myEntitlements(token).optJSONArray("entitlements") ?: JSONArray())
                }

                mainHandler.post {
                    loadingLabel.visibility = View.GONE
                    toolsContainer.removeAllViews()
                    if (tools.length() == 0) {
                        loadingLabel.text = "暂无可用工具"
                        loadingLabel.visibility = View.VISIBLE
                        return@post
                    }
                    for (i in 0 until tools.length()) {
                        val tool = tools.getJSONObject(i)
                        val owned = entitledToolCodes.contains(tool.optString("code"))
                        toolsContainer.addView(toolCard(tool, owned), matchWrap(if (i == 0) 0 else 10))
                    }
                }
            } catch (e: ApiError) {
                mainHandler.post {
                    if (e.status == 401) UserManager.logout(this)
                    loadingLabel.text = e.message ?: "工具加载失败"
                    loadingLabel.setTextColor(Color.rgb(220, 38, 38))
                }
            } catch (_: Exception) {
                mainHandler.post {
                    loadingLabel.text = "暂时无法连接服务器"
                    loadingLabel.setTextColor(Color.rgb(220, 38, 38))
                }
            }
        }.start()
    }

    private fun toolCard(tool: JSONObject, owned: Boolean): View {
        val name = tool.optString("name").ifBlank { tool.optString("code") }
        val desc = tool.optString("description")
        val card = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(dp(16), dp(16), dp(16), dp(14))
            background = cardBackground()
            elevation = dp(2).toFloat()
            isClickable = true
            isFocusable = true
            setOnClickListener { showToolDetail(tool, owned) }
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
        card.addView(TextView(this).apply {
            text = "查看详情  >"
            textSize = 13f
            setTextColor(Color.rgb(37, 99, 235))
            setPadding(0, dp(12), 0, 0)
        }, matchWrap())
        return card
    }

    private fun showToolDetail(tool: JSONObject, owned: Boolean) {
        currentTab = 1
        activeToolDetail = tool
        activeToolOwned = owned
        contentFrame.removeAllViews()
        val scroll = scrollable {
            setPadding(dp(20), dp(18), dp(20), dp(28))
        }
        val container = scroll.content
        val toolCode = tool.optString("code")
        val toolName = tool.optString("name").ifBlank { "自动点击器" }

        container.addView(TextView(this).apply {
            text = "<  工具中心"
            textSize = 14f
            setTextColor(Color.rgb(37, 99, 235))
            setPadding(0, dp(4), 0, dp(16))
            setOnClickListener { switchTab(1) }
        }, matchWrap())

        val header = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(dp(18), dp(18), dp(18), dp(16))
            background = cardBackground()
        }
        val titleRow = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
        }
        titleRow.addView(TextView(this).apply {
            text = toolName
            textSize = 21f
            setTextColor(Color.rgb(15, 23, 42))
        }, LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f))
        titleRow.addView(TextView(this).apply {
            text = if (owned) "已开通" else "未开通"
            textSize = 12f
            setTextColor(if (owned) Color.rgb(21, 128, 61) else Color.rgb(154, 103, 0))
            setPadding(dp(9), dp(4), dp(9), dp(4))
            background = GradientDrawable().apply {
                cornerRadius = dp(6).toFloat()
                setColor(if (owned) Color.rgb(232, 245, 241) else Color.rgb(255, 244, 214))
            }
        })
        header.addView(titleRow, matchWrap())
        header.addView(TextView(this).apply {
            text = tool.optString("description").ifBlank { "锁定屏幕坐标，按设定间隔持续点击。" }
            textSize = 13f
            setTextColor(Color.rgb(100, 116, 139))
            setPadding(0, dp(8), 0, 0)
        }, matchWrap())
        container.addView(header, matchWrap())

        if (!owned) {
            container.addView(sectionTitle("工具开通"), matchWrap(20))
            val accessPanel = LinearLayout(this).apply {
                orientation = LinearLayout.VERTICAL
                setPadding(dp(16), dp(15), dp(16), dp(16))
                background = GradientDrawable().apply {
                    cornerRadius = dp(8).toFloat()
                    setColor(Color.rgb(255, 250, 235))
                    setStroke(dp(1), Color.rgb(245, 201, 105))
                }
            }
            accessPanel.addView(TextView(this).apply {
                text = if (UserManager.isLoggedIn(this@MainActivity)) {
                    "当前账户尚未开通此工具，请输入购买的授权码。"
                } else {
                    "登录后才能兑换授权码并使用此工具。"
                }
                textSize = 13f
                setTextColor(Color.rgb(146, 89, 0))
            }, matchWrap())

            if (UserManager.isLoggedIn(this)) {
                val codeInput = EditText(this).apply {
                    hint = "AT-XXXX-XXXX-XXXX-XXXX"
                    textSize = 14f
                    isSingleLine = true
                    setPadding(dp(12), dp(11), dp(12), dp(11))
                    background = whiteInput()
                }
                accessPanel.addView(codeInput, matchWrap(14))
                val redeemButton = actionButton("兑换并开通")
                accessPanel.addView(redeemButton, matchWrap(10))
                redeemButton.setOnClickListener {
                    val licenseCode = codeInput.text.toString().trim()
                    if (licenseCode.isBlank()) {
                        Toast.makeText(this, "请输入授权码", Toast.LENGTH_SHORT).show()
                        return@setOnClickListener
                    }
                    redeemButton.isEnabled = false
                    redeemButton.text = "正在兑换..."
                    UserManager.redeemCode(this, licenseCode) { success, message, result ->
                        redeemButton.isEnabled = true
                        redeemButton.text = "兑换并开通"
                        if (!success) {
                            Toast.makeText(this, message, Toast.LENGTH_LONG).show()
                            return@redeemCode
                        }
                        val redeemedTool = result?.optJSONObject("tool")?.optString("code").orEmpty()
                        if (redeemedTool == toolCode) {
                            entitledToolCodes = entitledToolCodes + toolCode
                            activeToolOwned = true
                            Toast.makeText(this, "兑换成功，工具已开通", Toast.LENGTH_SHORT).show()
                            showToolDetail(tool, true)
                        } else {
                            Toast.makeText(this, "授权码已兑换，但对应的是其他工具", Toast.LENGTH_LONG).show()
                            switchTab(1)
                        }
                    }
                }
            } else {
                accessPanel.addView(actionButton("前往登录").apply {
                    setOnClickListener { switchTab(2) }
                }, matchWrap(14))
            }
            container.addView(accessPanel, matchWrap(8))
        }

        container.addView(sectionTitle("使用方法"), matchWrap(20))
        val guide = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(dp(16), dp(14), dp(16), dp(14))
            background = cardBackground()
        }
        listOf(
            "1. 开启悬浮窗权限和 AutomaticTools 无障碍服务。",
            "2. 启动悬浮控制，将准星拖到需要点击的位置。",
            "3. 点击“锁定”，选择或自定义点击间隔，再点击“开始”。",
            "4. 运行时可在迷你面板暂停、继续或展开设置。",
        ).forEachIndexed { index, text ->
            guide.addView(TextView(this).apply {
                this.text = text
                textSize = 13f
                setTextColor(Color.rgb(51, 65, 85))
                setPadding(0, if (index == 0) 0 else dp(10), 0, 0)
            }, matchWrap())
        }
        container.addView(guide, matchWrap(8))

        if (owned) {
            container.addView(sectionTitle("使用准备"), matchWrap(20))
            overlayStatus = statusText()
            accessibilityStatus = statusText()
            container.addView(permissionRow(
                "悬浮窗权限",
                overlayStatus,
                Settings.canDrawOverlays(this),
            ) { openOverlaySettings() }, matchWrap(8))
            container.addView(permissionRow(
                "无障碍服务",
                accessibilityStatus,
                ClickAccessibilityService.isRunning,
            ) { startActivity(Intent(Settings.ACTION_ACCESSIBILITY_SETTINGS)) }, matchWrap(10))

            container.addView(actionButton("启动悬浮控制").apply {
                setOnClickListener {
                    when {
                        !Settings.canDrawOverlays(this@MainActivity) ->
                            Toast.makeText(this@MainActivity, "请先开启悬浮窗权限", Toast.LENGTH_SHORT).show()
                        !ClickAccessibilityService.isRunning ->
                            Toast.makeText(this@MainActivity, "请先开启无障碍服务", Toast.LENGTH_SHORT).show()
                        else -> {
                            Toast.makeText(this@MainActivity, "正在验证工具授权...", Toast.LENGTH_SHORT).show()
                            startService(Intent(this@MainActivity, OverlayService::class.java))
                        }
                    }
                    refreshStatuses()
                }
            }, matchWrap(18))
            refreshStatuses()
        }

        contentFrame.addView(scroll.root)
    }

    private fun permissionRow(
        title: String,
        statusView: TextView,
        enabled: Boolean,
        openSettings: () -> Unit,
    ): View {
        return LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(dp(15), dp(12), dp(10), dp(12))
            background = cardBackground()

            val textColumn = LinearLayout(this@MainActivity).apply {
                orientation = LinearLayout.VERTICAL
                addView(TextView(this@MainActivity).apply {
                    text = title
                    textSize = 14f
                    setTextColor(Color.rgb(30, 41, 59))
                })
                addView(statusView.apply {
                    textSize = 12f
                    setTextColor(if (enabled) Color.rgb(22, 163, 74) else Color.rgb(148, 163, 184))
                    setPadding(0, dp(3), 0, 0)
                })
            }
            addView(textColumn, LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f))
            addView(Button(this@MainActivity).apply {
                text = if (enabled) "已开启" else "去设置"
                textSize = 12f
                isEnabled = !enabled
                setOnClickListener { openSettings() }
            }, LinearLayout.LayoutParams(LinearLayout.LayoutParams.WRAP_CONTENT, dp(42)))
        }
    }

    private fun sectionTitle(title: String): View = TextView(this).apply {
        text = title
        textSize = 16f
        setTextColor(Color.rgb(15, 23, 42))
    }

    private fun entitlementCodes(items: JSONArray): Set<String> {
        val result = mutableSetOf<String>()
        for (index in 0 until items.length()) {
            items.optJSONObject(index)?.optString("toolCode")?.takeIf { it.isNotBlank() }?.let(result::add)
        }
        return result
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
        overlayStatus.text = if (Settings.canDrawOverlays(this)) "已开启" else "未开启"
        accessibilityStatus.text = if (ClickAccessibilityService.isRunning) "已开启" else "未开启"
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

    private fun cardBackground(): GradientDrawable = GradientDrawable().apply {
        setColor(Color.WHITE)
        cornerRadius = dp(8).toFloat()
        setStroke(dp(1), Color.rgb(226, 232, 240))
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
