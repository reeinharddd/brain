#!/bin/bash
# ═══════════════════════════════════════════════════════════
#  brain/scripts/deploy.sh - Complete deployment automation
#  Usage: curl -fsSL https://raw.githubusercontent.com/username/brain/main/scripts/deploy.sh | bash
# ═══════════════════════════════════════════════════════════

set -euo pipefail

BRAIN_DIR="${BRAIN_DIR:-$HOME/.brain}"
REPO_URL="https://github.com/reeinharrrd/brain.git"
OS="unknown"
DEPLOYMENT_TYPE="minimal"

# ── Colors ───────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BLUE='\033[0;34m'; BOLD='\033[1m'; RESET='\033[0m'

ok()   { echo -e "  ${GREEN}✓${RESET} $1"; }
warn() { echo -e "  ${YELLOW}⚠${RESET}  $1"; }
fail() { echo -e "  ${RED}✗${RESET} $1"; }
info() { echo -e "  ${BLUE}→${RESET} $1"; }
section() { echo -e "\n${BOLD}── $1${RESET}"; }

# ── Usage ─────────────────────────────────────────────────────
usage() {
  echo "Brain Deployment Script"
  echo ""
  echo "Usage: $0 [options]"
  echo ""
  echo "Options:"
  echo "  --minimal     Deploy core MCPs only (default)"
  echo "  --full        Deploy all MCPs including GitHub, Slack, etc."
  echo "  --skills      Deploy with skills as MCP servers"
  echo "  --all        Deploy everything (full + skills)"
  echo "  --help        Show this help"
  echo ""
  echo "Environment Variables:"
  echo "  GITHUB_TOKEN       GitHub API token (for full deployment)"
  echo "  SLACK_BOT_TOKEN   Slack bot token (for full deployment)"
  echo "  TAVILY_API_KEY    Tavily API key (for skills deployment)"
  echo ""
  echo "Examples:"
  echo "  $0 --minimal"
  echo "  $0 --full"
  echo "  $0 --all"
}

# ── Parse Arguments ───────────────────────────────────────────
parse_args() {
  while [[ $# -gt 0 ]]; do
    case $1 in
      --minimal) DEPLOYMENT_TYPE="minimal"; shift ;;
      --full) DEPLOYMENT_TYPE="full"; shift ;;
      --skills) DEPLOYMENT_TYPE="skills"; shift ;;
      --all) DEPLOYMENT_TYPE="all"; shift ;;
      --help) usage; exit 0 ;;
      *) fail "Unknown option: $1"; usage; exit 1 ;;
    esac
  done
}

# ── Environment Detection ─────────────────────────────────────
detect_env() {
  section "Detecting Environment"
  
  # OS Detection
  if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    if grep -q Microsoft /proc/version 2>/dev/null; then
      OS="wsl"; info "Windows Subsystem for Linux (WSL)"
    else
      OS="linux"; info "Linux"
    fi
  elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macos"; info "macOS"
  else
    warn "Unknown OS: $OSTYPE"
  fi
  
  # Check dependencies
  if command -v docker &>/dev/null; then ok "Docker found"; else fail "Docker required"; exit 1; fi
  if command -v git &>/dev/null; then ok "Git found"; else fail "Git required"; exit 1; fi
  if command -v node &>/dev/null; then ok "Node.js found"; else warn "Node.js not found (will be installed)"; fi
  
  info "Deployment type: $DEPLOYMENT_TYPE"
}

# ── Install Dependencies ───────────────────────────────────────
install_deps() {
  section "Installing Dependencies"
  
  # Install Docker if missing (Linux only)
  if ! command -v docker &>/dev/null && [[ "$OS" == "linux" ]]; then
    info "Installing Docker..."
    curl -fsSL https://get.docker.com | sh
    sudo usermod -aG docker "$USER"
    ok "Docker installed (relogin required)"
  fi
  
  # Install Node.js if missing
  if ! command -v node &>/dev/null; then
    info "Installing Node.js..."
    if [[ "$OS" == "macos" ]]; then
      brew install node
    else
      curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
      sudo apt-get install -y nodejs
    fi
    ok "Node.js installed"
  fi
}

# ── Clone/Update Brain Repo ────────────────────────────────────
setup_brain_repo() {
  section "Setting up Brain Repository"
  
  if [ -d "$BRAIN_DIR/.git" ]; then
    info "Updating existing brain repo..."
    cd "$BRAIN_DIR"
    git pull origin main
    ok "Brain repo updated"
  else
    info "Cloning brain repo..."
    git clone "$REPO_URL" "$BRAIN_DIR"
    ok "Brain repo cloned"
  fi
  
  # Run install script
  cd "$BRAIN_DIR"
  bash scripts/install.sh
  ok "Brain repo installed"
}

# ── Configure Environment ───────────────────────────────────────
configure_env() {
  section "Configuring Environment"
  
  local env_file="$BRAIN_DIR/docker/.env"
  
  # Create .env from template if missing
  if [ ! -f "$env_file" ]; then
    cp "$BRAIN_DIR/docker/.env.example" "$env_file"
    ok ".env created from template"
  fi
  
  # Set HOST_HOME
  sed -i "s|HOST_HOME=.*|HOST_HOME=$HOME|" "$env_file"
  
  # Add additional ports for full deployment
  if [[ "$DEPLOYMENT_TYPE" == "full" ]] || [[ "$DEPLOYMENT_TYPE" == "all" ]]; then
    if ! grep -q "MCP_GITHUB_PORT" "$env_file"; then
      cat >> "$env_file" << EOF

# Extended MCP ports
MCP_GITHUB_PORT=3004
MCP_POSTGRES_PORT=3006
MCP_SLACK_PORT=3007
MCP_LINEAR_PORT=3008
MCP_RESEARCH_PORT=3009
MCP_FRONTEND_PORT=3010
EOF
      ok "Extended ports configured"
    fi
  fi
  
  # Check for required tokens
  if [[ "$DEPLOYMENT_TYPE" == "full" ]] || [[ "$DEPLOYMENT_TYPE" == "all" ]]; then
    if [[ -z "${GITHUB_TOKEN:-}" ]]; then
      warn "GITHUB_TOKEN not set - GitHub MCP will not work"
    fi
    if [[ -z "${SLACK_BOT_TOKEN:-}" ]]; then
      warn "SLACK_BOT_TOKEN not set - Slack MCP will not work"
    fi
  fi
  
  if [[ "$DEPLOYMENT_TYPE" == "skills" ]] || [[ "$DEPLOYMENT_TYPE" == "all" ]]; then
    if [[ -z "${TAVILY_API_KEY:-}" ]]; then
      warn "TAVILY_API_KEY not set - Research skill will not work"
    fi
  fi
}

# ── Deploy Services ─────────────────────────────────────────────
deploy_services() {
  section "Deploying MCP Services"
  
  cd "$BRAIN_DIR/docker"
  
  case $DEPLOYMENT_TYPE in
    minimal)
      info "Starting minimal MCP stack..."
      docker compose up -d
      ;;
    full)
      info "Starting full MCP stack..."
      docker compose --profile extended up -d
      ;;
    skills)
      info "Starting MCP stack with skills..."
      docker compose --profile skills up -d
      ;;
    all)
      info "Starting complete MCP stack with all services..."
      docker compose --profile extended --profile skills up -d
      ;;
  esac
  
  ok "Services deployed"
}

# ── Setup Skills as MCPs ───────────────────────────────────────
setup_skills() {
  if [[ "$DEPLOYMENT_TYPE" != "skills" ]] && [[ "$DEPLOYMENT_TYPE" != "all" ]]; then
    return
  fi
  
  section "Setting up Skills as MCP Servers"
  
  local skills_dir="$BRAIN_DIR/docker/skills"
  mkdir -p "$skills_dir"
  
  # Create MCP server wrappers for skills
  cat > "$skills_dir/research-server.js" << 'EOF'
// Research Skill MCP Server
const { Server } = require('@modelcontextprotocol/sdk/server/index.js');
const { StdioServerTransport } = require('@modelcontextprotocol/sdk/server/stdio.js');

const server = new Server(
  {
    name: "research-skill",
    version: "1.0.0",
  },
  {
    capabilities: {
      tools: {},
    },
  }
);

server.setRequestHandler("tools/list", async () => {
  return {
    tools: [
      {
        name: "research",
        description: "Comprehensive research with web data and citations",
        inputSchema: {
          type: "object",
          properties: {
            input: { type: "string", description: "Research topic or question" },
            model: { type: "string", enum: ["mini", "pro", "auto"], default: "mini" }
          },
          required: ["input"]
        }
      }
    ]
  };
});

server.setRequestHandler("tools/call", async (request) => {
  const { name, arguments: args } = request.params;
  
  if (name === "research") {
    // Call the actual research skill script
    const { spawn } = require('child_process');
    const result = await new Promise((resolve, reject) => {
      const proc = spawn('/home/$(process.env.USER)/.claude/skills/research/scripts/research.sh', 
        [JSON.stringify(args)], 
        { stdio: 'pipe' }
      );
      
      let output = '';
      proc.stdout.on('data', (data) => output += data.toString());
      proc.on('close', (code) => {
        if (code === 0) resolve({ content: [{ type: "text", text: output }] });
        else reject(new Error(`Research failed with code ${code}`));
      });
    });
    
    return result;
  }
  
  throw new Error(`Unknown tool: ${name}`);
});

async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error("Research MCP Server running on stdio");
}

main().catch(console.error);
EOF
  
  ok "Research skill MCP server created"
  
  # TODO: Add more skill servers as needed
  warn "Additional skills need to be implemented as MCP servers"
}

# ── Verify Deployment ───────────────────────────────────────────
verify_deployment() {
  section "Verifying Deployment"
  
  cd "$BRAIN_DIR/docker"
  
  # Wait for services to start
  info "Waiting for services to start..."
  sleep 10
  
  # Check core services
  local ports=(3001 3002 3003 3005)
  local names=("memory" "filesystem" "sequential" "context7")
  
  if [[ "$DEPLOYMENT_TYPE" == "full" ]] || [[ "$DEPLOYMENT_TYPE" == "all" ]]; then
    ports+=(3004)
    names+=("github")
  fi
  
  if [[ "$DEPLOYMENT_TYPE" == "skills" ]] || [[ "$DEPLOYMENT_TYPE" == "all" ]]; then
    ports+=(3009)
    names+=("research")
  fi
  
  local failed=0
  for i in "${!ports[@]}"; do
    local port="${ports[$i]}"
    local name="${names[$i]}"
    
    if curl -sf "http://localhost:$port/sse" &>/dev/null; then
      ok "$name (port $port) responding"
    else
      fail "$name (port $port) not responding"
      ((failed++))
    fi
  done
  
  if [ $failed -eq 0 ]; then
    ok "All services verified successfully"
  else
    warn "$failed services failed to start"
  fi
}

# ── Show Next Steps ─────────────────────────────────────────────
show_next_steps() {
  section "Deployment Complete"
  
  echo -e "\n${BOLD}🎉 Brain MCP Stack deployed successfully!${RESET}"
  echo ""
  echo -e "${BOLD}Active Services:${RESET}"
  
  case $DEPLOYMENT_TYPE in
    minimal)
      echo "  • Memory (port 3001)"
      echo "  • Filesystem (port 3002)"
      echo "  • Sequential Thinking (port 3003)"
      echo "  • Context7 (port 3005)"
      ;;
    full)
      echo "  • All minimal services"
      echo "  • GitHub (port 3004)"
      ;;
    skills)
      echo "  • All minimal services"
      echo "  • Research Skill (port 3009)"
      ;;
    all)
      echo "  • All minimal services"
      echo "  • GitHub (port 3004)"
      echo "  • Research Skill (port 3009)"
      ;;
  esac
  
  echo ""
  echo -e "${BOLD}Next Steps:${RESET}"
  echo "  1. Configure your AI client to use MCP endpoints:"
  echo "     Claude Code: Already configured via Docker settings"
  echo "     Other clients: Use http://localhost:300X/sse URLs"
  echo ""
  echo "  2. Set up tokens for extended services:"
  echo "     export GITHUB_TOKEN=your_token"
  echo "     export SLACK_BOT_TOKEN=your_token"
  echo "     export TAVILY_API_KEY=your_token"
  echo ""
  echo "  3. Manage services:"
  echo "     Stop:   ~/.brain/docker/start.sh down"
  echo "     Status: ~/.brain/docker/start.sh status"
  echo "     Logs:   ~/.brain/docker/start.sh logs"
  echo ""
  
  ok "Deployment complete!"
}

# ── Main ─────────────────────────────────────────────────────
main() {
  parse_args "$@"
  detect_env
  install_deps
  setup_brain_repo
  configure_env
  setup_skills
  deploy_services
  verify_deployment
  show_next_steps
}

# Execute if run directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi
