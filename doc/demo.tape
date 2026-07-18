# DarkSuitAI README demo recorder.
#
# Produces doc/demo.gif — the animation shown at the top of the README.
#
# Prereqs:
#   1. Install VHS:  brew install vhs   (or: go install github.com/charmbracelet/vhs@latest)
#   2. Put a working key in .env at the repo root (see .env.example).
#      Recommended for a clean demo:
#        DARKSUIT_PROVIDER=anthropic
#        ANTHROPIC_API_KEY=sk-ant-...
#
# Record it (from the repo root):
#   vhs doc/demo.tape
#
# Tip: if the model response runs long, bump the Sleep values below.

Output doc/demo.gif

Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 720
Set Theme "Dracula"
Set Padding 20
Set TypingSpeed 45ms

# --- intro ---
Type "# DarkSuitAI — production AI agents in pure Go"
Enter
Sleep 800ms
Type "# define a tool, run an agent, stream the answer:"
Enter
Sleep 800ms

# --- run the bundled example ---
Type "go run ./doc"
Enter

# Give it time to compile + make the live calls. Increase if needed.
Sleep 25s

# hold the final frame
Sleep 2s
