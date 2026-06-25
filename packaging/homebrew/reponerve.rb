class Reponerve < Formula
  desc "Local-first software understanding for developers and AI agents"
  homepage "https://github.com/reponerve/reponerve"
  version "1.4.0"
  license "Apache-2.0"

  on_macos do
    on_arm do
      url "https://github.com/reponerve/reponerve/releases/download/v1.4.0/reponerve_1.4.0_darwin_arm64.tar.gz"
      sha256 "REPLACE_ON_RELEASE"
    end
    on_intel do
      url "https://github.com/reponerve/reponerve/releases/download/v1.4.0/reponerve_1.4.0_darwin_amd64.tar.gz"
      sha256 "REPLACE_ON_RELEASE"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/reponerve/reponerve/releases/download/v1.4.0/reponerve_1.4.0_linux_arm64.tar.gz"
      sha256 "REPLACE_ON_RELEASE"
    end
    on_intel do
      url "https://github.com/reponerve/reponerve/releases/download/v1.4.0/reponerve_1.4.0_linux_amd64.tar.gz"
      sha256 "REPLACE_ON_RELEASE"
    end
  end

  def install
    bin.install "reponerve"
  end

  test do
    assert_match "RepoNerve", shell_output("#{bin}/reponerve --help")
  end
end
