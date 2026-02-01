# ⚙️ Variables
APP_NAME := tspider
MAIN_FILE := cmd/tspider/main.go
BUILD_DIR := bin

# Detect OS type for sed compatibility (Linux or macOS)
ifeq ($(shell uname), Darwin)
	SED_INPLACE = -i ''
else
	SED_INPLACE = -i
endif

# 🏗️ Build the project
build:
	@echo "🔨 Building..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags '-s -w' -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "✅ Build completed: $(BUILD_DIR)/$(APP_NAME)"

# 🚀 Run the project
run: build
	@echo "🏃 Running..."
	./$(BUILD_DIR)/$(APP_NAME)

# 🧪 Run tests
test:
	@echo "🧪 Running tests..."
	go test ./... -v
	@echo "✅ Tests completed!"

# 🩺 Run doctor check
doctor: build
	@echo "🩺 Running doctor check..."
	./$(BUILD_DIR)/$(APP_NAME) doctor

# 🧹 Clean build artifacts
clean:
	@echo "🧹 Cleaning build files..."
	rm -rf $(BUILD_DIR)
	@echo "✅ Clean completed!"

# 🏷️ Tag version, update main.go, and push to GitHub (tag should be in vX.X.X format with integers >= 0)
tag:
	@if [ -z "$(TAG)" ]; then \
		echo "❗ Usage: make tag TAG=vX.X.X"; \
		exit 1; \
	fi
	@if ! echo "$(strip $(TAG))" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$$'; then \
		echo "❌ Invalid tag format! Use vX.X.X format (e.g., v1.0.0, v0.2.3, v2.1.5)"; \
		exit 1; \
	fi
	@echo "🔄 Updating version in $(MAIN_FILE) to: $(TAG)"
	@sed $(SED_INPLACE) "s/^\(var version = \)\".*\"/\1\"$(TAG)\"/" $(MAIN_FILE)
	git add $(MAIN_FILE)
	git commit -m "🔖 Version update: $(TAG)"
	git tag -a $(TAG) -m "🔖 Release $(TAG)"
	@echo "⬆️ Pushing tag $(TAG) to GitHub..."
	git push origin $(TAG)
	@echo "⬆️ Pushing main branch to GitHub..."
	git push origin main
	@echo "✅ Tag $(TAG) and main branch updated on GitHub!"

# 🆘 Display help (default target)
help:
	@echo "📘 Available commands:"
	@echo "  make build              - Build the project"
	@echo "  make run                - Build and run the project"
	@echo "  make test               - Run all tests"
	@echo "  make doctor             - Check torrent site availability"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make tag TAG=vX.X.X     - Update version in main.go, create an annotated Git tag (vX.X.X format, integers >= 0), and push to GitHub (also updates main branch)"
