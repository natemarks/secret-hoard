#!/usr/bin/env bash
# Install git pre-commit hook to run 'make static' before commits
#
# Usage:
#   ./scripts/git-commithook.sh install   - Install the pre-commit hook
#   ./scripts/git-commithook.sh uninstall - Remove the pre-commit hook
#   ./scripts/git-commithook.sh status    - Check if hook is installed

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOOK_FILE="${REPO_ROOT}/.git/hooks/pre-commit"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

install_hook() {
    echo "Installing pre-commit hook..."

    # Check if .git directory exists
    if [ ! -d "${REPO_ROOT}/.git" ]; then
        echo -e "${RED}Error: Not a git repository${NC}"
        exit 1
    fi

    # Create hooks directory if it doesn't exist
    mkdir -p "${REPO_ROOT}/.git/hooks"

    # Check if hook already exists
    if [ -f "${HOOK_FILE}" ]; then
        echo -e "${YELLOW}Warning: pre-commit hook already exists${NC}"
        read -p "Overwrite? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Installation cancelled"
            exit 0
        fi
        # Backup existing hook
        cp "${HOOK_FILE}" "${HOOK_FILE}.backup.$(date +%Y%m%d%H%M%S)"
        echo "Backed up existing hook"
    fi

    # Create the pre-commit hook
    cat > "${HOOK_FILE}" << 'EOF'
#!/usr/bin/env bash
# Pre-commit hook: Run 'make static' before allowing commit
#
# This hook runs all static checks including:
# - Unit tests (no external dependencies)
# - Go imports formatting
# - Go formatting
# - Go vet
# - Linting
# - Cyclomatic complexity check
# - Dead code detection
# - Vulnerability check
# - All Go tests
#
# To bypass this hook (not recommended):
#   git commit --no-verify

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Running pre-commit checks (make static)...${NC}"
echo ""

# Run make static from repository root
REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "${REPO_ROOT}"

if make static; then
    echo ""
    echo -e "${GREEN}✓ All static checks passed${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}✗ Static checks failed${NC}"
    echo ""
    echo "Please fix the errors above before committing."
    echo "Or use 'git commit --no-verify' to bypass (not recommended)."
    exit 1
fi
EOF

    # Make the hook executable
    chmod +x "${HOOK_FILE}"

    echo -e "${GREEN}✓ Pre-commit hook installed successfully${NC}"
    echo ""
    echo "The hook will run 'make static' before every commit."
    echo "This includes:"
    echo "  • Unit tests (fast, no external dependencies)"
    echo "  • Code formatting checks"
    echo "  • Static analysis"
    echo "  • All Go tests"
    echo ""
    echo "To bypass the hook (not recommended):"
    echo "  git commit --no-verify"
}

uninstall_hook() {
    echo "Uninstalling pre-commit hook..."

    if [ ! -f "${HOOK_FILE}" ]; then
        echo -e "${YELLOW}No pre-commit hook found${NC}"
        exit 0
    fi

    # Check if it's our hook
    if grep -q "make static" "${HOOK_FILE}"; then
        rm "${HOOK_FILE}"
        echo -e "${GREEN}✓ Pre-commit hook uninstalled${NC}"
    else
        echo -e "${YELLOW}Warning: Hook exists but doesn't appear to be ours${NC}"
        echo "Contents:"
        head -5 "${HOOK_FILE}"
        echo ""
        read -p "Remove anyway? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm "${HOOK_FILE}"
            echo -e "${GREEN}✓ Pre-commit hook removed${NC}"
        else
            echo "Uninstall cancelled"
        fi
    fi
}

check_status() {
    echo "Checking pre-commit hook status..."
    echo ""

    if [ ! -d "${REPO_ROOT}/.git" ]; then
        echo -e "${RED}✗ Not a git repository${NC}"
        exit 1
    fi

    if [ -f "${HOOK_FILE}" ]; then
        if grep -q "make static" "${HOOK_FILE}"; then
            echo -e "${GREEN}✓ Pre-commit hook is installed${NC}"
            echo ""
            echo "Hook location: ${HOOK_FILE}"
            echo "Hook runs: make static"
            echo ""
            # Show first few lines of hook
            echo "Hook contents (first 10 lines):"
            head -10 "${HOOK_FILE}" | sed 's/^/  /'
        else
            echo -e "${YELLOW}⚠ A pre-commit hook exists but it's not ours${NC}"
            echo ""
            echo "Hook location: ${HOOK_FILE}"
            echo ""
            echo "Hook contents (first 10 lines):"
            head -10 "${HOOK_FILE}" | sed 's/^/  /'
        fi
    else
        echo -e "${YELLOW}✗ No pre-commit hook installed${NC}"
        echo ""
        echo "To install: ./scripts/git-commithook.sh install"
    fi
}

show_help() {
    cat << EOF
Git Pre-Commit Hook Installation Script

Usage:
  $0 <command>

Commands:
  install     Install the pre-commit hook
  uninstall   Remove the pre-commit hook
  status      Check if the hook is installed
  help        Show this help message

Description:
  Installs a git pre-commit hook that runs 'make static' before allowing
  commits. This ensures all static checks pass before code is committed.

  The hook includes:
    • Unit tests (no external dependencies, ~3ms)
    • Go imports formatting (goimports)
    • Go code formatting (fmt)
    • Go vet
    • Linting
    • Cyclomatic complexity check
    • Dead code detection
    • Vulnerability scanning
    • All Go tests

Examples:
  # Install the hook
  ./scripts/git-commithook.sh install

  # Check if installed
  ./scripts/git-commithook.sh status

  # Remove the hook
  ./scripts/git-commithook.sh uninstall

  # Bypass the hook for a single commit (not recommended)
  git commit --no-verify

Notes:
  - The hook backs up any existing pre-commit hook before overwriting
  - The hook runs from the repository root
  - The hook can be bypassed with 'git commit --no-verify'
  - Backups are saved as .git/hooks/pre-commit.backup.<timestamp>
EOF
}

# Main script logic
case "${1:-}" in
    install)
        install_hook
        ;;
    uninstall)
        uninstall_hook
        ;;
    status)
        check_status
        ;;
    help|--help|-h)
        show_help
        ;;
    "")
        echo -e "${RED}Error: No command specified${NC}"
        echo ""
        show_help
        exit 1
        ;;
    *)
        echo -e "${RED}Error: Unknown command '${1}'${NC}"
        echo ""
        show_help
        exit 1
        ;;
esac
