default: install

# Variables
home_dir := env_var('HOME')

# Install the Go application
install:
	templ generate
	go install -tags "fts5" .

mac-service-install: install
    mkdir -p "{{home_dir}}/Library/LaunchAgents"
    echo "Installing launch agent..."
    sed 's|$HOME|{{home_dir}}|g' template.plist > "{{home_dir}}/Library/LaunchAgents/com.lukaswerner.mark.server.plist"
    launchctl unload "{{home_dir}}/Library/LaunchAgents/com.lukaswerner.mark.server.plist" 2>/dev/null || true
    launchctl load "{{home_dir}}/Library/LaunchAgents/com.lukaswerner.mark.server.plist"
    echo "Launch agent installed and loaded"

# Uninstall the launch agent
mac-service-uninstall:
    launchctl unload "{{home_dir}}/Library/LaunchAgents/com.lukaswerner.mark.server.plist" 2>/dev/null || true
    rm -f "{{home_dir}}/Library/LaunchAgents/com.lukaswerner.mark.server.plist"
    echo "Launch agent uninstalled"

# Check status of the service
mac-status:
    launchctl list | grep "com.lukaswerner.mark.server" || echo "Service not running"

# View logs
mac-logs:
    echo "=== STDOUT ==="
    cat /tmp/mark.server.stdout.log 2>/dev/null || echo "No stdout log found"
    echo -e "\n=== STDERR ==="
    cat /tmp/mark.server.stderr.log 2>/dev/null || echo "No stderr log found"

# Restart the service
mac-restart: mac-service-uninstall mac-service-install

# Clean up everything
clean:
    rm -f {{home_dir}}/.lib/crsqlite.dylib
