class Gns3util < Formula
  desc "GNS3 API utility for managing GNS3v3 servers"
  homepage "https://github.com/Stefanistkuhl/gns3-api-util"
  url "https://github.com/Stefanistkuhl/gns3-api-util/releases/download/v1.0.1/gns3util-darwin-amd64.tar.gz"
  sha256 "PLACEHOLDER_SHA256"
  license "GPL-3.0-or-later"
  version "1.0.1"

  # Detect architecture
  if Hardware::CPU.arm?
    url "https://github.com/Stefanistkuhl/gns3-api-util/releases/download/v1.0.1/gns3util-darwin-arm64.tar.gz"
    sha256 "PLACEHOLDER_SHA256_ARM64"
  end

  def install
    bin.install "gns3util"
    
    # Install shell completions if they exist
    if File.exist?("completions/gns3util.bash")
      bash_completion.install "completions/gns3util.bash" => "gns3util"
    end
    if File.exist?("completions/_gns3util")
      zsh_completion.install "completions/_gns3util"
    end
    if File.exist?("completions/gns3util.fish")
      fish_completion.install "completions/gns3util.fish"
    end
    
    # Install man pages if they exist
    man1.install "man/gns3util.1" if File.exist?("man/gns3util.1")
  end

  test do
    system "#{bin}/gns3util", "--help"
  end
end
