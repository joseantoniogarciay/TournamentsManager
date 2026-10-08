package expo.modules.globalfeedback

import android.content.Context
import expo.modules.kotlin.AppContext
import expo.modules.kotlin.functions.Queues
import expo.modules.kotlin.modules.Module
import expo.modules.kotlin.modules.ModuleDefinition
import expo.modules.kotlin.records.Field
import expo.modules.kotlin.records.Record
import expo.modules.kotlin.views.ExpoView

class BannerSpec : Record {
  @Field var id: Int = 0
  @Field var message: String = ""
  @Field var expiresAt: Double = 0.0
  @Field var backgroundColor: String = ""
  @Field var borderColor: String = ""
  @Field var textColor: String = ""
  @Field var fontFamily: String = ""
  @Field var fontSize: Double = 0.0
  @Field var lineHeight: Double = 0.0
  @Field var padding: Double = 0.0
  @Field var borderRadius: Double = 0.0
  @Field var borderWidth: Double = 0.0
  @Field var left: Double = 0.0
  @Field var top: Double = 0.0
  @Field var width: Double = 0.0
  @Field var enterExitMs: Int = 0
  @Field var feedbackMs: Int = 0
  @Field var swipeDistance: Double = 0.0
  @Field var swipeSlop: Double = 0.0
  @Field var reducedMotion: Boolean = false
}

// One controller per Expo runtime, shared by the root and native window anchors.
internal object BannerControllers {
  private val controllers = mutableMapOf<AppContext, BannerController>()
  fun get(context: AppContext) = controllers.getOrPut(context) { BannerController(context) }
  fun destroy(context: AppContext) { controllers.remove(context)?.destroy() }
}

class FeedbackWindowAnchor(context: Context, appContext: AppContext) : ExpoView(context, appContext) {
  private val controller = BannerControllers.get(appContext)
  override fun onAttachedToWindow() {
    super.onAttachedToWindow()
    controller.attach(this)
  }

  override fun onDetachedFromWindow() {
    controller.detach(this)
    super.onDetachedFromWindow()
  }
}

class GlobalFeedbackModule : Module() {
  override fun definition() = ModuleDefinition {
    Name("TMGlobalFeedback")
    Events("onDismiss")

    OnCreate {
      BannerControllers.get(appContext).onDismiss = { id -> sendEvent("onDismiss", mapOf("id" to id)) }
    }
    AsyncFunction("show") { spec: BannerSpec -> BannerControllers.get(appContext).show(spec) }
      .runOnQueue(Queues.MAIN)
    AsyncFunction("hide") { reducedMotion: Boolean, duration: Int ->
      BannerControllers.get(appContext).hide(reducedMotion, duration)
    }.runOnQueue(Queues.MAIN)
    OnActivityEntersBackground { BannerControllers.get(appContext).pause() }
    OnActivityEntersForeground { BannerControllers.get(appContext).resume() }
    OnActivityDestroys { BannerControllers.get(appContext).activityDestroyed() }
    OnDestroy { BannerControllers.destroy(appContext) }

    View(FeedbackWindowAnchor::class) {}
  }
}
