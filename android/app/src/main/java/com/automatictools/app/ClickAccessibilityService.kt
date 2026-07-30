package com.automatictools.app

import android.accessibilityservice.AccessibilityService
import android.accessibilityservice.GestureDescription
import android.graphics.Path
import android.os.Handler
import android.os.Looper
import android.view.accessibility.AccessibilityEvent

class ClickAccessibilityService : AccessibilityService() {
    override fun onServiceConnected() {
        instance = this
        isRunning = true
    }

    override fun onDestroy() {
        if (instance === this) {
            instance = null
        }
        isRunning = false
        super.onDestroy()
    }

    override fun onAccessibilityEvent(event: AccessibilityEvent?) = Unit

    override fun onInterrupt() = Unit

    private fun tap(x: Int, y: Int, callback: (Boolean) -> Unit) {
        val path = Path().apply {
            moveTo(x.toFloat(), y.toFloat())
        }
        val gesture = GestureDescription.Builder()
            .addStroke(GestureDescription.StrokeDescription(path, 0, 60))
            .build()

        dispatchGesture(
            gesture,
            object : GestureResultCallback() {
                override fun onCompleted(gestureDescription: GestureDescription?) {
                    callback(true)
                }

                override fun onCancelled(gestureDescription: GestureDescription?) {
                    callback(false)
                }
            },
            Handler(Looper.getMainLooper()),
        )
    }

    companion object {
        @Volatile
        var isRunning: Boolean = false
            private set

        @Volatile
        private var instance: ClickAccessibilityService? = null

        fun performTap(x: Int, y: Int, callback: (Boolean) -> Unit): Boolean {
            val service = instance ?: return false
            service.tap(x, y, callback)
            return true
        }
    }
}
