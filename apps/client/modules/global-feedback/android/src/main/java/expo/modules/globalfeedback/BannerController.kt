package expo.modules.globalfeedback

import android.annotation.SuppressLint
import android.graphics.Color
import android.graphics.PixelFormat
import android.graphics.Typeface
import android.graphics.drawable.GradientDrawable
import android.os.Handler
import android.os.Looper
import android.util.TypedValue
import android.view.Gravity
import android.view.MotionEvent
import android.view.VelocityTracker
import android.view.View
import android.view.WindowManager
import android.widget.TextView
import androidx.core.widget.TextViewCompat
import com.facebook.react.common.assets.ReactFontManager
import expo.modules.kotlin.AppContext
import kotlin.math.abs
import kotlin.math.min
import kotlin.math.roundToInt

internal class BannerController(private val appContext: AppContext) {
  var onDismiss: ((Int) -> Unit)? = null
  private val handler = Handler(Looper.getMainLooper())
  private val anchors = linkedSetOf<View>()
  private var spec: BannerSpec? = null
  private var windowManager: WindowManager? = null
  private var text: TextView? = null
  private var attachedWindow: View? = null
  private var foreground = true
  private var destroyed = false

  fun attach(anchor: View) {
    anchors.add(anchor)
    // The window token is available after attachment; keep the same notice and deadline.
    anchor.post { if (!destroyed && anchor in anchors) present(false) }
  }

  fun detach(anchor: View) {
    anchors.remove(anchor)
    if (attachedWindow === anchor) {
      removeWindow()
      attachedWindow = null
      handler.post { if (!destroyed) present(false) }
    }
  }

  fun show(next: BannerSpec) {
    val isNew = spec?.id != next.id
    spec = next
    present(isNew)
  }

  fun hide(reducedMotion: Boolean, duration: Int) {
    spec = null
    val label = text ?: return
    label.animate().cancel()
    if (reducedMotion || text?.parent == null) {
      removeWindow()
      attachedWindow = null
    } else {
      label.animate().alpha(0f).setDuration(duration.toLong()).withEndAction {
        if (spec == null) {
          removeWindow()
          attachedWindow = null
        }
      }.start()
    }
  }

  fun pause() {
    foreground = false
    removeWindow()
    attachedWindow = null
  }

  fun resume() {
    foreground = true
    present(false)
  }

  fun activityDestroyed() {
    pause()
    text?.animate()?.cancel()
    windowManager = null
    text = null
    anchors.clear()
  }

  fun destroy() {
    destroyed = true
    handler.post {
      spec = null
      text?.animate()?.cancel()
      removeWindow()
      windowManager = null
      text = null
      anchors.clear()
      attachedWindow = null
      onDismiss = null
    }
  }

  private fun present(animate: Boolean) {
    val current = spec ?: return
    if (!foreground || destroyed) return
    if (current.expiresAt <= System.currentTimeMillis()) {
      removeWindow()
      attachedWindow = null
      return
    }
    val anchor = anchors.lastOrNull { it.isAttachedToWindow && it.windowToken != null } ?: return
    val activity = appContext.currentActivity ?: return
    if (activity.isFinishing || activity.isDestroyed) return
    val density = anchor.resources.displayMetrics.density
    fun dp(value: Double) = (value * density).roundToInt()
    val label = text ?: TextView(activity).also { text = it; installGestures(it) }
    label.animate().cancel()
    label.alpha = 1f
    label.translationY = 0f
    label.text = current.message
    label.setTextColor(Color.parseColor(current.textColor))
    label.setTextSize(TypedValue.COMPLEX_UNIT_SP, current.fontSize.toFloat())
    label.typeface = ReactFontManager.getInstance().getTypeface(current.fontFamily, Typeface.NORMAL, activity.assets)
    TextViewCompat.setLineHeight(label, (current.lineHeight * label.resources.displayMetrics.scaledDensity).roundToInt())
    label.setPadding(dp(current.padding), dp(current.padding), dp(current.padding), dp(current.padding))
    label.background = GradientDrawable().apply {
      setColor(Color.parseColor(current.backgroundColor))
      cornerRadius = dp(current.borderRadius).toFloat()
      setStroke(dp(current.borderWidth).coerceAtLeast(1), Color.parseColor(current.borderColor))
    }
    label.isClickable = true
    label.importantForAccessibility = View.IMPORTANT_FOR_ACCESSIBILITY_YES
    label.accessibilityLiveRegion = View.ACCESSIBILITY_LIVE_REGION_POLITE

    val params = WindowManager.LayoutParams(
      dp(current.width),
      WindowManager.LayoutParams.WRAP_CONTENT,
      WindowManager.LayoutParams.TYPE_APPLICATION_SUB_PANEL,
      WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE or
        WindowManager.LayoutParams.FLAG_NOT_TOUCH_MODAL or
        WindowManager.LayoutParams.FLAG_LAYOUT_IN_SCREEN,
      PixelFormat.TRANSLUCENT,
    ).apply {
      token = anchor.windowToken
      gravity = Gravity.TOP or Gravity.LEFT
      x = dp(current.left)
      y = dp(current.top)
      // Window titles can be exposed by accessibility services as well.
      setTitle(current.message)
    }
    if (attachedWindow !== anchor || label.parent == null) {
      removeWindow()
      val manager = activity.getSystemService(WindowManager::class.java)
      windowManager = manager
      try {
        manager.addView(label, params)
        attachedWindow = anchor
      } catch (_: WindowManager.BadTokenException) {
        // A dialog may lose its token while its anchor is detaching. The next
        // attached anchor presents the same notice, with the original deadline.
        attachedWindow = null
        return
      }
    } else {
      windowManager?.updateViewLayout(label, params)
    }
    if (animate && !current.reducedMotion) {
      label.alpha = 0f
      label.translationY = -dp(current.padding).toFloat()
      label.animate().alpha(1f).translationY(0f).setDuration(current.enterExitMs.toLong()).start()
    }
  }

  private fun removeWindow() {
    attachedWindow = null
    val label = text ?: return
    if (label.parent != null) {
      try {
        windowManager?.removeViewImmediate(label)
      } catch (_: IllegalArgumentException) {
        // WindowManager may have already removed a destroyed parent window.
      }
    }
  }

  @SuppressLint("ClickableViewAccessibility")
  private fun installGestures(label: TextView) {
    var startX = 0f
    var startY = 0f
    var dragging = false
    var gestureId: Int? = null
    var velocity: VelocityTracker? = null
    label.setOnClickListener { spec?.let { onDismiss?.invoke(it.id) } }
    label.setOnTouchListener { _, event ->
      val current = spec ?: return@setOnTouchListener false
      val density = label.resources.displayMetrics.density
      if (event.actionMasked != MotionEvent.ACTION_DOWN && gestureId != current.id) {
        velocity?.recycle()
        velocity = null
        dragging = false
        label.translationY = 0f
        return@setOnTouchListener true
      }
      when (event.actionMasked) {
        MotionEvent.ACTION_DOWN -> {
          gestureId = current.id
          label.animate().cancel()
          startX = event.rawX
          startY = event.rawY
          dragging = false
          velocity?.recycle()
          velocity = VelocityTracker.obtain().also { it.addMovement(event) }
        }
        MotionEvent.ACTION_MOVE -> {
          velocity?.addMovement(event)
          val dy = event.rawY - startY
          if (dy < -current.swipeSlop * density && abs(dy) > abs(event.rawX - startX)) dragging = true
          if (dragging) label.translationY = min(dy, 0f)
        }
        MotionEvent.ACTION_UP, MotionEvent.ACTION_CANCEL -> {
          velocity?.addMovement(event)
          velocity?.computeCurrentVelocity(1000)
          val dismiss = event.actionMasked == MotionEvent.ACTION_UP && dragging &&
            (event.rawY - startY <= -current.swipeDistance * density || (velocity?.yVelocity ?: 0f) < -500 * density)
          velocity?.recycle()
          velocity = null
          if (dismiss) onDismiss?.invoke(current.id)
          else if (!dragging && event.actionMasked == MotionEvent.ACTION_UP) label.performClick()
          else label.animate().translationY(0f).setDuration(if (current.reducedMotion) 0 else current.feedbackMs.toLong()).start()
        }
      }
      true
    }
  }
}
