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
      sha256 "d068b26ae065abd550e898d1218a68fe2fade757f463c2a76f9f0995923981e3" # darwin-arm64
    else
      url "https://github.com/imcclaskey/d3/releases/download/v#{version}/d3-darwin-amd64"
      sha256 "94496dd4ab9134a80ca0bc7a975c69c6b135249546583df9c057afe17afb5b5c" # darwin-amd64
    end
  end
  
  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/imcclaskey/d3/releases/download/v#{version}/d3-linux-arm64"
      sha256 "aa4e6e5f2f0d73111e91ebf31c676d6c47a657b493dbbbecc072ab850c6c4ba3" # linux-arm64
    else
      url "https://github.com/imcclaskey/d3/releases/download/v#{version}/d3-linux-amd64"
      sha256 "798086483bafc4d1a4376ada3ea2e8846413a817a51cda4432822e542c498598" # linux-amd64
    end
  end
  
  def install
    bin.install Dir["d3*"].first => "d3"
  end
  
  test do
    assert_match version.to_s, shell_output("#{bin}/d3 --version")
  end
end 