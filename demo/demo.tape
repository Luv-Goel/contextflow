# ContextFlow demo — record with: vhs demo/demo.tape
# Install VHS: https://github.com/charmbracelet/vhs

Output demo/demo.gif

Set Theme "Dracula"
Set FontSize 14
Set Width 1000
Set Height 600
Set Padding 20

# Setup: install cf
Type "# Install ContextFlow"
Enter
Sleep 500ms
Type "curl -fsSL https://raw.githubusercontent.com/Luv-Goel/contextflow/main/scripts/install.sh | bash"
Sleep 500ms
# Simulated — don't actually run in CI
Ctrl+C
Sleep 500ms

Type "# Add to your shell (already done in this demo)"
Enter
Sleep 300ms
Type "# eval \"$(cf init zsh)\"  ← add this to ~/.zshrc"
Enter
Sleep 800ms

# Show version
Type "cf version"
Enter
Sleep 800ms

# Show help
Type "cf --help"
Enter
Sleep 1500ms

# Simulate some commands having been recorded
Type "# After using your shell normally for a while..."
Enter
Sleep 500ms

# Search demo
Type "# Search with Ctrl+R — fuzzy search your entire history"
Enter
Sleep 800ms
Type "cf search docker"
Enter
Sleep 2000ms
Escape
Sleep 500ms

# Workflows demo
Type "# Browse auto-detected workflows"
Enter
Sleep 300ms
Type "cf workflows"
Enter
Sleep 2000ms
Escape
Sleep 500ms

# Stats demo
Type "# Usage statistics"
Enter
Sleep 300ms
Type "cf stats"
Enter
Sleep 1500ms

# Export demo
Type "# Export workflow #1 as a shell script"
Enter
Sleep 300ms
Type "cf export 1 > setup.sh && cat setup.sh"
Enter
Sleep 1500ms

# Tag demo
Type "# Give a workflow a custom name"
Enter
Sleep 300ms
Type "cf tag 1 \"docker-dev-setup\""
Enter
Sleep 1000ms

Type "# That's ContextFlow. Never forget a workflow again. 🧠"
Enter
Sleep 2000ms
