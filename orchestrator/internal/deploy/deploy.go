package deploy

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
)

// UploadDirAsTarGz uploads localDir as remoteDir/bundle.tar.gz and ensures remoteDir exists.
func UploadDirAsTarGz(s *sftp.Client, localDir, remoteDir string) error {
	if err := s.MkdirAll(remoteDir); err != nil {
		// Ignore if it already exists
		_ = err
	}

	remotePath := filepath.ToSlash(filepath.Join(remoteDir, "bundle.tar.gz"))
	dst, err := s.Create(remotePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	gw := gzip.NewWriter(dst)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(localDir, func(path string, info os.FileInfo, werr error) error {
		if werr != nil {
			return werr
		}
		if info.IsDir() {
			return nil
		}

		rel, _ := filepath.Rel(localDir, path)
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(rel)

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		_, err = io.Copy(tw, src)
		return err
	})
}
