class D3 < Formula
  desc "D3 - a tool for software development workflows"
  homepage "https://github.com/imcclaskey/d3"
  
  # Use the version from the constant in version.go
  # The version here should match the constant in internal/version/version.go
  version "0.2.0"
  
  # Alternatively, you would need to update this version manually when releasing
  # The best practice is to keep this in sync with the Go code
  
  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/imcclaskey/d3/releases/download/v#{version}/d3-darwin-arm64"
      sha256 "60d4e3d777652d0242ecaa0ab3e70db2c99e485f399535c94fb6d5b7db23b714" # darwin-arm64
    else
      url "https://github.com/imcclaskey/d3/releases/download/v#{version}/d3-darwin-amd64"
      sha256 "40b57cda3c8c3cdd238ce6e620cab786a85cc6f1ff16108f770d56de4b2a9e00" # darwin-amd64
    end
  end
  
  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/imcclaskey/d3/releases/download/v#{version}/d3-linux-arm64"
      sha256 "6abc3de5a4df76b5b6acc39d6cd74fc940490bbc3a78451e699e6b82a63ea251" # linux-arm64
    else
      url "https://github.com/imcclaskey/d3/releases/download/v#{version}/d3-linux-amd64"
      sha256 "a3511f9bc94daeb437291e1e1e677fc23ccfe29f0cd9cc74189fea1421084937" # linux-amd64
    end
  end
  
  def install
    bin.install Dir["d3*"].first => "d3"
  end
  
  test do
    assert_match version.to_s, shell_output("#{bin}/d3 --version")
  end
end 