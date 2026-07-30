package com.automatictools.app

import android.app.Activity
import android.app.AlertDialog
import android.content.Intent
import android.graphics.Color
import android.graphics.drawable.GradientDrawable
import android.net.Uri
import android.os.Bundle
import android.provider.Settings
import android.view.Gravity
import android.view.View
import android.widget.Button
import android.widget.EditText
import android.widget.LinearLayout
import android.widget.TextView
import android.widget.Toast

class MainActivity : Activity() {
    private lateinit var overlayStatus: TextView
    private lateinit var accessibilityStatus: TextView
    private lateinit var contentFrame: LinearLayout
    
    private var currentTab = 0 // 0: Home, 1: Personal

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

    private fun buildRootView(): View {
        val root = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setBackgroundColor(Color.rgb(239, 246, 255))
        }

        contentFrame = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            layoutParams = LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT,
                0,
                1f
            )
        }
        root.addView(contentFrame)

        // Bottom Navigation Bar
        val navBar = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            setBackgroundColor(Color.WHITE)
            elevation = dp(8).toFloat()
            setPadding(0, dp(8), 0, dp(8))
        }

        val homeTab = navItem("主页", 0)
        val profileTab = navItem("个人", 1)
        
        navBar.addView(homeTab, LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f))
        navBar.addView(profileTab, LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f))
        
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
        
        // Update Bottom Nav Colors (re-building nav is overkill, but simple for this script)
        // For a more robust solution we'd keep references to the nav textviews
        val navBar = (contentFrame.parent as LinearLayout).getChildAt(1) as LinearLayout
        for (i in 0 until navBar.childCount) {
            val tv = navBar.getChildAt(i) as TextView
            tv.setTextColor(if (i == index) Color.rgb(37, 99, 235) else Color.GRAY)
        }

        if (index == 0) {
            contentFrame.addView(buildHomeView())
            refreshStatuses()
        } else {
            contentFrame.addView(buildPersonalView())
        }
    }

    private fun buildHomeView(): View {
        val container = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(dp(24), dp(28), dp(24), dp(24))
        }

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
        }, matchWrap(top = 18))

        container.addView(actionButton("开启无障碍服务").apply {
            setOnClickListener {
                startActivity(Intent(Settings.ACTION_ACCESSIBILITY_SETTINGS))
            }
        }, matchWrap(top = 10))

        container.addView(actionButton("启动悬浮控制").apply {
            setOnClickListener {
                if (Settings.canDrawOverlays(this@MainActivity)) {
                    startService(Intent(this@MainActivity, OverlayService::class.java))
                } else {
                    Toast.makeText(this@MainActivity, "请先开启悬浮窗权限", Toast.LENGTH_SHORT).show()
                }
                refreshStatuses()
            }
        }, matchWrap(top = 18))

        container.addView(TextView(this).apply {
            text = "拖动准星到目标位置，点“锁定”，再点“开始”。运行后面板会自动收起。"
            textSize = 13f
            setTextColor(Color.rgb(71, 85, 105))
            setPadding(0, dp(18), 0, 0)
        }, matchWrap())

        return container
    }

    private fun buildPersonalView(): View {
        val container = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            gravity = Gravity.CENTER_HORIZONTAL
            setPadding(dp(24), dp(40), dp(24), dp(24))
        }

        val isLoggedIn = UserManager.isLoggedIn(this)
        val username = UserManager.getUsername(this) ?: "未登录用户"

        // Avatar Placeholder
        container.addView(View(this).apply {
            layoutParams = LinearLayout.LayoutParams(dp(80), dp(80)).apply {
                bottomMargin = dp(16)
            }
            background = GradientDrawable().apply {
                shape = GradientDrawable.OVAL
                setColor(if (isLoggedIn) Color.rgb(37, 99, 235) else Color.rgb(226, 232, 240))
            }
        })

        container.addView(TextView(this).apply {
            text = if (isLoggedIn) username else "未登录用户"
            textSize = 20f
            setTextColor(Color.rgb(15, 23, 42))
            setPadding(0, 0, 0, dp(8))
        }, matchWrap())

        container.addView(TextView(this).apply {
            text = if (isLoggedIn) "欢迎回来，点击同步云端配置" else "登录后同步配置信息"
            textSize = 14f
            setTextColor(Color.rgb(100, 116, 139))
            setPadding(0, 0, 0, dp(32))
            gravity = Gravity.CENTER
        }, matchWrap())

        if (!isLoggedIn) {
            container.addView(actionButton("立即登录 / 注册").apply {
                setOnClickListener { showLoginDialog() }
            }, matchWrap())
        } else {
            container.addView(actionButton("退出登录").apply {
                background = GradientDrawable().apply {
                    cornerRadius = dp(12).toFloat()
                    setColor(Color.rgb(220, 38, 38))
                }
                setOnClickListener {
                    UserManager.logout(this@MainActivity)
                    switchTab(1)
                }
            }, matchWrap())
        }

        // Some list items
        val menuList = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(0, dp(40), 0, 0)
        }

        menuList.addView(menuItem("我的配置"))
        menuList.addView(menuItem("使用教程"))
        menuList.addView(menuItem("关于我们"))

        container.addView(menuList, matchWrap())

        return container
    }

    private fun showLoginDialog() {
        val layout = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(dp(20), dp(20), dp(20), dp(20))
        }

        val userEdit = EditText(this).apply { hint = "用户名" }
        val passEdit = EditText(this).apply { hint = "密码"; inputType = android.text.InputType.TYPE_CLASS_TEXT or android.text.InputType.TYPE_TEXT_VARIATION_PASSWORD }
        
        layout.addView(userEdit)
        layout.addView(passEdit)

        AlertDialog.Builder(this)
            .setTitle("登录 / 注册")
            .setView(layout)
            .setPositiveButton("登录") { _, _ ->
                val u = userEdit.text.toString()
                val p = passEdit.text.toString()
                UserManager.login(this, u, p) { success, msg ->
                    Toast.makeText(this, msg, Toast.LENGTH_SHORT).show()
                    if (success) switchTab(1)
                }
            }
            .setNeutralButton("注册") { _, _ ->
                val u = userEdit.text.toString()
                val p = passEdit.text.toString()
                if (UserManager.register(this, u, p)) {
                    Toast.makeText(this, "注册成功，请登录", Toast.LENGTH_SHORT).show()
                }
            }
            .setNegativeButton("取消", null)
            .show()
    }

    private fun menuItem(title: String): View = LinearLayout(this).apply {
        orientation = LinearLayout.HORIZONTAL
        setPadding(dp(16), dp(16), dp(16), dp(16))
        background = GradientDrawable().apply {
            setColor(Color.WHITE)
            cornerRadius = dp(8).toFloat()
        }
        layoutParams = LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.MATCH_PARENT,
            LinearLayout.LayoutParams.WRAP_CONTENT
        ).apply { topMargin = dp(8) }

        addView(TextView(this@MainActivity).apply {
            text = title
            textSize = 15f
            setTextColor(Color.rgb(30, 41, 59))
        })
    }

    private fun refreshStatuses() {
        overlayStatus.text = if (Settings.canDrawOverlays(this)) {
            "悬浮窗权限：已开启"
        } else {
            "悬浮窗权限：未开启"
        }

        accessibilityStatus.text = if (ClickAccessibilityService.isRunning) {
            "无障碍服务：已开启"
        } else {
            "无障碍服务：未开启"
        }
    }

    private fun openOverlaySettings() {
        val intent = Intent(
            Settings.ACTION_MANAGE_OVERLAY_PERMISSION,
            Uri.parse("package:$packageName"),
        )
        startActivity(intent)
    }

    private fun statusText(): TextView = TextView(this).apply {
        textSize = 15f
        setTextColor(Color.rgb(30, 41, 59))
        setPadding(0, dp(4), 0, dp(4))
    }

    private fun actionButton(textValue: String): Button =
        Button(this).apply {
            text = textValue
            textSize = 15f
            setTextColor(Color.WHITE)
            background = GradientDrawable().apply {
                cornerRadius = dp(12).toFloat()
                setColor(Color.rgb(37, 99, 235))
            }
        }

    private fun matchWrap(top: Int = 0): LinearLayout.LayoutParams =
        LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.MATCH_PARENT,
            LinearLayout.LayoutParams.WRAP_CONTENT,
        ).apply {
            topMargin = dp(top)
        }

    private fun dp(value: Int): Int =
        (value * resources.displayMetrics.density).toInt()
}
