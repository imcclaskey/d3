class D3 < Formula
  desc "D3 - a tool for software development workflows"
  homepage "https://github.com/imcclaskey/d3"
  
  # Use the version from the constant in version.go
  # The version here should match the constant in internal/version/version.go
  version "0.1.0"
  
  # Alternatively, you would need to update this version manually when releasing
  # The best practice is to keep this in sync with the Go code
  
  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/imcclaskey/d3/releases/download/v#{version}/d3-darwin-arm64"
      sha256 "a96096a1250d2bd4d219c0ca38fc7930fa9b6e0c953b6d4f9ffbabe1f352ba3a" # darwin-arm64
    else
      url "https://github.com/imcclaskey/d3/releases/download/v#{version}/d3-darwin-amd64"
      sha256 "bb5385dc5ab814808806953976038dd49b61a0f3de24e2b65d1a3b0471e4fd56" # darwin-amd64
    end
  end
  
  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/imcclaskey/d3/releases/download/v#{version}/d3-linux-arm64"
      sha256 "c9619b4d985cf070cbf55505a3be712a57e37dcced871e4c0c6f0f26a80df1c3" # linux-arm64
    else
      url "https://github.com/imcclaskey/d3/releases/download/v#{version}/d3-linux-amd64"
      sha256 "e41c4aa5e6a51f435fdc91d447397294b9bd5ce44af7acf39e8d9c0e40d99cf1" # linux-amd64
    end
  end
  
  def install
    bin.install Dir["d3*"].first => "d3"
  end
  
  test do
    assert_match version.to_s, shell_output("#{bin}/d3 --version")
  end
end 