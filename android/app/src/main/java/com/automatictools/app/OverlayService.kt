package com.automatictools.app

import android.app.Service
import android.content.Intent
import android.graphics.Color
import android.graphics.PixelFormat
import android.graphics.drawable.GradientDrawable
import android.os.Handler
import android.os.IBinder
import android.os.Looper
import android.view.Gravity
import android.view.MotionEvent
import android.view.View
import android.view.WindowManager
import android.widget.Button
import android.widget.LinearLayout
import android.widget.TextView
import android.widget.Toast

class OverlayService : Service() {
    private val targetSizeDp = 32
    private val blue = Color.rgb(37, 99, 235)
    private val cyan = Color.rgb(14, 165, 233)
    private val red = Color.rgb(220, 38, 38)
    private val ink = Color.rgb(15, 23, 42)
    private val muted = Color.rgb(71, 85, 105)
    private val surface = Color.rgb(248, 250, 252)

    private lateinit var windowManager: WindowManager
    private lateinit var targetView: TextView
    private lateinit var panelView: LinearLayout
    private lateinit var miniView: LinearLayout
    private lateinit var targetParams: WindowManager.LayoutParams
    private lateinit var panelParams: WindowManager.LayoutParams
    private lateinit var miniParams: WindowManager.LayoutParams
    private lateinit var coordinateText: TextView
    private lateinit var statusText: TextView
    private lateinit var countText: TextView
    private lateinit var miniCountText: TextView
    private lateinit var miniStopBtn: TextView
    private lateinit var intervalText: TextView
    private lateinit var startButton: Button

    private val handler = Handler(Looper.getMainLooper())
    private var lockedX: Int? = null
    private var lockedY: Int? = null
    private var running = false
    private var targetAttached = false
    private var panelAttached = false
    private var miniAttached = false
    private var intervalMs = 1000L
    private var clickCount = 0

    override fun onCreate() {
        super.onCreate()
        windowManager = getSystemService(WINDOW_SERVICE) as WindowManager
        createTarget()
        createPanel()
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        expandPanel()
        return START_STICKY
    }

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onDestroy() {
        stopClicking()
        if (::targetView.isInitialized && targetAttached) {
            windowManager.removeView(targetView)
            targetAttached = false
        }
        if (::panelView.isInitialized && panelAttached) {
            windowManager.removeView(panelView)
            panelAttached = false
        }
        if (::miniView.isInitialized && miniAttached) {
            windowManager.removeView(miniView)
            miniAttached = false
        }
        super.onDestroy()
    }

    private fun createTarget() {
        targetView = TextView(this).apply {
            text = "+"
            textSize = 22f
            gravity = Gravity.CENTER
            setTextColor(Color.WHITE)
            background = GradientDrawable().apply {
                shape = GradientDrawable.OVAL
                setColor(Color.argb(190, 14, 165, 233))
                setStroke(dp(2), Color.WHITE)
            }
        }

        targetParams = overlayParams(dp(targetSizeDp), dp(targetSizeDp)).apply {
            x = dp(320)
            y = dp(360)
        }

        targetView.setOnTouchListener(DragTouchListener(targetParams) {
            windowManager.updateViewLayout(targetView, targetParams)
            updateCoordinateText()
        })

        windowManager.addView(targetView, targetParams)
        targetAttached = true
    }

    private fun createPanel() {
        panelView = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(0, 0, 0, dp(10))
            background = GradientDrawable().apply {
                cornerRadius = dp(12).toFloat()
                setColor(surface)
                setStroke(dp(1), Color.rgb(191, 219, 254))
            }
        }

        val titleLayout = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            background = GradientDrawable().apply {
                cornerRadii = floatArrayOf(
                    dp(12).toFloat(), dp(12).toFloat(),
                    dp(12).toFloat(), dp(12).toFloat(),
                    0f, 0f,
                    0f, 0f,
                )
                setColor(blue)
            }
        }

        val titleText = TextView(this).apply {
            text = " 拖动这里 · 点击器"
            textSize = 13f
            includeFontPadding = false
            setTextColor(Color.WHITE)
            gravity = Gravity.CENTER_VERTICAL
            setPadding(dp(12), 0, dp(0), 0)
        }
        titleLayout.addView(titleText, LinearLayout.LayoutParams(0, dp(40), 1f))

        val shrinkBtn = TextView(this).apply {
            text = "收起"
            textSize = 12f
            includeFontPadding = false
            setTextColor(Color.WHITE)
            setPadding(dp(12), 0, dp(15), 0)
            gravity = Gravity.CENTER
            setOnClickListener { collapsePanel() }
        }
        titleLayout.addView(shrinkBtn, LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.WRAP_CONTENT, dp(40)
        ))

        panelView.addView(titleLayout, matchWrap())

        val content = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(dp(10), dp(8), dp(10), 0)
        }
        panelView.addView(content, matchWrap())

        coordinateText = label("目标：-, -")
        statusText = label("状态：停止")
        countText = label("点击次数：0")
        intervalText = label("间隔：1000ms")
        content.addView(coordinateText, matchWrap())
        content.addView(statusText, matchWrap())
        content.addView(countText, matchWrap())
        content.addView(intervalText, matchWrap())

        content.addView(intervalRow(), matchWrap(top = 8))
        content.addView(actionRow(), matchWrap(top = 8))

        startButton = panelButton("开始", Color.rgb(22, 163, 74)).apply {
            setOnClickListener { toggleClicking() }
        }
        content.addView(startButton, matchWrap(top = 8))

        panelParams = overlayParams(dp(270), WindowManager.LayoutParams.WRAP_CONTENT).apply {
            x = dp(16)
            y = dp(24)
        }

        titleLayout.setOnTouchListener(DragTouchListener(panelParams) {
            windowManager.updateViewLayout(panelView, panelParams)
        })

        windowManager.addView(panelView, panelParams)
        panelAttached = true
        updateCoordinateText()
    }

    private fun createMiniPanel() {
        miniView = LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            gravity = Gravity.CENTER_VERTICAL
            setPadding(dp(8), dp(5), dp(8), dp(5))
            background = GradientDrawable().apply {
                cornerRadius = dp(18).toFloat()
                setColor(Color.argb(235, 15, 23, 42))
                setStroke(dp(1), Color.argb(210, 125, 211, 252))
            }
        }

        miniCountText = TextView(this).apply {
            text = "0 次"
            textSize = 13f
            setTextColor(Color.WHITE)
            setPadding(dp(8), dp(4), dp(8), dp(4))
            background = GradientDrawable().apply {
                cornerRadius = dp(14).toFloat()
                setColor(cyan)
            }
        }
        miniView.addView(
            miniCountText,
            LinearLayout.LayoutParams(LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT),
        )

        miniStopBtn = TextView(this).apply {
            text = "停止"
            textSize = 13f
            gravity = Gravity.CENTER
            setTextColor(Color.WHITE)
            setPadding(dp(12), dp(4), dp(12), dp(4))
            background = GradientDrawable().apply {
                cornerRadius = dp(14).toFloat()
                setColor(red)
            }
            setOnClickListener { stopClicking() }
        }
        miniView.addView(miniStopBtn, LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT
        ).apply { leftMargin = dp(6) })

        val expandBtn = TextView(this).apply {
            text = "展开"
            textSize = 13f
            gravity = Gravity.CENTER
            setTextColor(Color.WHITE)
            setPadding(dp(12), dp(4), dp(12), dp(4))
            background = GradientDrawable().apply {
                cornerRadius = dp(14).toFloat()
                setStroke(dp(1), Color.WHITE)
            }
            setOnClickListener { expandPanel() }
        }
        miniView.addView(expandBtn, LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.WRAP_CONTENT, LinearLayout.LayoutParams.WRAP_CONTENT
        ).apply { leftMargin = dp(6) })

        miniParams = overlayParams(WindowManager.LayoutParams.WRAP_CONTENT, WindowManager.LayoutParams.WRAP_CONTENT).apply {
            x = panelParams.x
            y = panelParams.y
        }

        miniView.setOnTouchListener(DragTouchListener(miniParams) {
            windowManager.updateViewLayout(miniView, miniParams)
        })
    }

    private fun intervalRow(): LinearLayout {
        return LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL
            addView(intervalButton("100", 100), rowButtonParams(0, 4))
            addView(intervalButton("500", 500), rowButtonParams(4, 4))
            addView(intervalButton("1000", 1000), rowButtonParams(4, 4))
            addView(intervalButton("2秒", 2000), rowButtonParams(4, 0))
        }
    }

    private fun actionRow(): LinearLayout {
        return LinearLayout(this).apply {
            orientation = LinearLayout.HORIZONTAL

            addView(panelButton("锁定", blue).apply {
                setOnClickListener { lockTarget() }
            }, rowButtonParams(0, 4))

            addView(panelButton("清零", Color.rgb(245, 158, 11)).apply {
                setOnClickListener {
                    clickCount = 0
                    updateCountText()
                }
            }, rowButtonParams(4, 4))

            addView(panelButton("关闭", Color.rgb(100, 116, 139)).apply {
                setOnClickListener { stopSelf() }
            }, rowButtonParams(4, 0))
        }
    }

    private fun intervalButton(label: String, value: Long): Button {
        return panelButton(label, Color.rgb(226, 232, 240), darkText = true).apply {
            minWidth = 0
            setOnClickListener {
                intervalMs = value
                intervalText.text = "间隔：${formatInterval(value)}"
            }
        }
    }

    private fun lockTarget() {
        val center = targetCenter()
        lockedX = center.first
        lockedY = center.second
        coordinateText.text = "已锁定：${center.first}, ${center.second}"
        statusText.text = "状态：已锁定"
    }

    private fun toggleClicking() {
        if (running) {
            stopClicking()
        } else {
            startClicking()
        }
    }

    private fun startClicking() {
        val x = lockedX
        val y = lockedY
        if (x == null || y == null) {
            Toast.makeText(this, "请先拖动准星并点击“锁定”。", Toast.LENGTH_SHORT).show()
            return
        }
        if (!ClickAccessibilityService.isRunning) {
            Toast.makeText(this, "请先开启 AutomaticTools 无障碍服务。", Toast.LENGTH_LONG).show()
            return
        }

        running = true
        detachTargetView()
        collapsePanel()
        startButton.text = "停止"
        statusText.text = "状态：点击中"
        scheduleClick()
    }

    private fun stopClicking() {
        running = false
        handler.removeCallbacksAndMessages(null)
        attachTargetView()
        expandPanel()
        if (::startButton.isInitialized) {
            startButton.text = "开始"
        }
        if (::statusText.isInitialized) {
            statusText.text = if (lockedX != null) "状态：已锁定" else "状态：停止"
        }
    }

    private fun scheduleClick() {
        if (!running) return
        val x = lockedX ?: return
        val y = lockedY ?: return

        val dispatched = ClickAccessibilityService.performTap(x, y) { success ->
            handler.post {
                if (success) {
                    clickCount += 1
                    updateCountText()
                } else {
                    statusText.text = "状态：点击被取消"
                }
                if (running) {
                    handler.postDelayed({ scheduleClick() }, intervalMs)
                }
            }
        }

        if (!dispatched) {
            running = false
            startButton.text = "开始"
            statusText.text = "状态：需要无障碍服务"
            attachTargetView()
            expandPanel()
        }
    }

    private fun updateCoordinateText() {
        if (!::coordinateText.isInitialized) return
        val center = targetCenter()
        if (lockedX == null || !running) {
            coordinateText.text = if (lockedX == null) {
                "目标：${center.first}, ${center.second}"
            } else {
                "已锁定：$lockedX, $lockedY"
            }
        }
    }

    private fun updateCountText() {
        countText.text = "点击次数：$clickCount"
        if (::miniCountText.isInitialized) {
            miniCountText.text = "$clickCount 次"
        }
    }

    private fun targetCenter(): Pair<Int, Int> {
        if (::targetView.isInitialized && targetAttached && targetView.width > 0 && targetView.height > 0) {
            val location = IntArray(2)
            targetView.getLocationOnScreen(location)
            return Pair(location[0] + targetView.width / 2, location[1] + targetView.height / 2)
        }

        val size = dp(targetSizeDp)
        return Pair(targetParams.x + size / 2, targetParams.y + size / 2)
    }

    private fun detachTargetView() {
        if (targetAttached) {
            windowManager.removeView(targetView)
            targetAttached = false
        }
    }

    private fun attachTargetView() {
        if (::targetView.isInitialized && !targetAttached) {
            windowManager.addView(targetView, targetParams)
            targetAttached = true
            updateCoordinateText()
        }
    }

    private fun collapsePanel() {
        if (!::miniView.isInitialized) {
            createMiniPanel()
        }
        miniStopBtn.visibility = if (running) View.VISIBLE else View.GONE
        if (panelAttached) {
            miniParams.x = panelParams.x
            miniParams.y = panelParams.y
            windowManager.removeView(panelView)
            panelAttached = false
        }
        miniCountText.text = "$clickCount 次"
        if (!miniAttached) {
            windowManager.addView(miniView, miniParams)
            miniAttached = true
        }
    }

    private fun expandPanel() {
        if (::miniView.isInitialized && miniAttached) {
            panelParams.x = miniParams.x
            panelParams.y = miniParams.y
            windowManager.removeView(miniView)
            miniAttached = false
        }
        if (::panelView.isInitialized && !panelAttached) {
            windowManager.addView(panelView, panelParams)
            panelAttached = true
        }
    }

    private fun label(textValue: String): TextView =
        TextView(this).apply {
            text = textValue
            textSize = 14f
            setTextColor(muted)
            setPadding(0, dp(2), 0, dp(2))
        }

    private fun panelButton(textValue: String, color: Int, darkText: Boolean = false): Button =
        Button(this).apply {
            text = textValue
            textSize = 13f
            minHeight = dp(36)
            setPadding(dp(4), 0, dp(4), 0)
            setTextColor(if (darkText) ink else Color.WHITE)
            background = GradientDrawable().apply {
                cornerRadius = dp(12).toFloat()
                setColor(color)
            }
        }

    private fun formatInterval(value: Long): String =
        if (value >= 1000 && value % 1000L == 0L) {
            "${value / 1000}秒"
        } else {
            "${value}ms"
        }

    private fun overlayParams(width: Int, height: Int): WindowManager.LayoutParams =
        WindowManager.LayoutParams(
            width,
            height,
            WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY,
            WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE,
            PixelFormat.TRANSLUCENT,
        ).apply {
            gravity = Gravity.TOP or Gravity.START
        }

    private fun matchWrap(top: Int = 0): LinearLayout.LayoutParams =
        LinearLayout.LayoutParams(
            LinearLayout.LayoutParams.MATCH_PARENT,
            LinearLayout.LayoutParams.WRAP_CONTENT,
        ).apply {
            topMargin = dp(top)
        }

    private fun rowButtonParams(left: Int, right: Int): LinearLayout.LayoutParams =
        LinearLayout.LayoutParams(0, LinearLayout.LayoutParams.WRAP_CONTENT, 1f).apply {
            leftMargin = dp(left)
            rightMargin = dp(right)
        }

    private inner class DragTouchListener(
        private val params: WindowManager.LayoutParams,
        private val onMove: () -> Unit,
    ) : View.OnTouchListener {
        private var startX = 0
        private var startY = 0
        private var downRawX = 0f
        private var downRawY = 0f

        override fun onTouch(view: View, event: MotionEvent): Boolean {
            when (event.action) {
                MotionEvent.ACTION_DOWN -> {
                    startX = params.x
                    startY = params.y
                    downRawX = event.rawX
                    downRawY = event.rawY
                    return true
                }
                MotionEvent.ACTION_MOVE -> {
                    params.x = startX + (event.rawX - downRawX).toInt()
                    params.y = startY + (event.rawY - downRawY).toInt()
                    onMove()
                    return true
                }
            }
            return true
        }
    }

    private fun dp(value: Int): Int =
        (value * resources.displayMetrics.density).toInt()
}
