# AutomaticTools Android

Native Kotlin Android client for AutomaticTools.

## Features

- Floating draggable target dot.
- Floating control panel.
- Lock target coordinate.
- Start and stop auto tapping.
- Interval buttons: 100ms, 500ms, 1000ms, 2s.
- Click counter and reset button.
- Uses AccessibilityService gestures, no root required.

## Build

Open this folder in Android Studio:

```text
D:\wcs\Code\AutomaticTools\android
```

Then run:

```text
Build > Build APK(s)
```

For the website download package, build the signed release APK:

```powershell
.\gradlew.bat assembleRelease
```

The release signing identity is stored locally in `keystore/` and
`keystore.properties`. Both paths are ignored by Git. Back up both files
together; losing this key prevents future APK versions from upgrading the
installed release app.

The development APK and release APK use different signatures. Uninstall an
existing development build once before installing the first release build.
Future releases must keep this signing key and increment `versionCode`.

The current machine does not expose `java` or `gradle` in PATH, so this project was not compiled in this Codex run.

## Phone Setup

1. Install the APK.
2. Open AutomaticTools.
3. Tap `Open overlay permission` and allow display over other apps.
4. Tap `Open accessibility settings`.
5. Enable the `AutomaticTools` accessibility service.
6. Return to AutomaticTools and tap `Start floating controls`.

## Usage

1. Drag the blue target dot to the target tap position.
2. Tap `Lock`.
3. Pick an interval.
4. Tap `Start`.
5. Tap `Stop` to pause.
6. Tap `Reset` to clear the click count.

Some apps, especially games, payment apps, banking apps, and anti-automation protected apps, may block accessibility generated taps.
