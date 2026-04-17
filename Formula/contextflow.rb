# typed: false
# frozen_string_literal: true

# Homebrew formula for ContextFlow.
# To publish: create a tap repo Luv-Goel/homebrew-tap and put this file there.
# Then users can: brew install Luv-Goel/tap/contextflow
#
# SHA256 values are populated automatically by the release workflow.
# Until v0.1.0 is tagged, this is a template.

class Contextflow < Formula
  desc "Shell history that understands your workflows, not just your commands"
  homepage "https://github.com/Luv-Goel/contextflow"
  license "MIT"
  version "0.1.0"

  on_macos do
    on_arm do
      url "https://github.com/Luv-Goel/contextflow/releases/download/v#{version}/cf-darwin-arm64.tar.gz"
      sha256 "PLACEHOLDER_DARWIN_ARM64_SHA256"
    end
    on_intel do
      url "https://github.com/Luv-Goel/contextflow/releases/download/v#{version}/cf-darwin-amd64.tar.gz"
      sha256 "PLACEHOLDER_DARWIN_AMD64_SHA256"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Luv-Goel/contextflow/releases/download/v#{version}/cf-linux-arm64.tar.gz"
      sha256 "PLACEHOLDER_LINUX_ARM64_SHA256"
    end
    on_intel do
      url "https://github.com/Luv-Goel/contextflow/releases/download/v#{version}/cf-linux-amd64.tar.gz"
      sha256 "PLACEHOLDER_LINUX_AMD64_SHA256"
    end
  end

  def install
    bin.install Dir["cf-*"].first => "cf"
  end

  def caveats
    <<~EOS
      To enable shell integration, add to your shell config:

        bash (~/.bashrc):
          eval "$(cf init bash)"

        zsh (~/.zshrc):
          eval "$(cf init zsh)"

        fish (~/.config/fish/config.fish):
          cf init fish | source

      Then restart your shell and press Ctrl+R to search.
    EOS
  end

  test do
    assert_match "cf version #{version}", shell_output("#{bin}/cf version")
  end
end
