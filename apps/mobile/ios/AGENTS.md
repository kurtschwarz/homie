# iOS app

Use the mobile MCP tools to verify changes that affect the running iOS app. Start with `mobile_list_available_devices` and use a suitable local simulator when available. Launch the built app with `mobile_launch_app`, inspect screens with `mobile_list_elements_on_screen`, and exercise the changed flow. Use screenshots only when hierarchy inspection cannot answer a visual question. Report the device and the behavior verified.

For build or test verification, use the existing Xcode project and test targets. Pair build/test results with an on-device or simulator check whenever the change affects app behavior or UI.
