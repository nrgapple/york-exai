package app

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PersistedArtifact struct {
	RelativePath string
	AbsolutePath string
	OriginalName string
	SHA256       string
}

type BackupManifest struct {
	CreatedAt     string `json:"created_at"`
	SchemaVersion int    `json:"schema_version"`
	DatabasePath  string `json:"database_path"`
	ArtifactRoot  string `json:"artifact_root"`
	ArtifactCount int    `json:"artifact_count"`
}

func (r *Runtime) PersistArtifact(src string, kind string, ownerID string) (PersistedArtifact, error) {
	targetDir, err := artifactDirForKind(r.Paths, kind, ownerID)
	if err != nil {
		return PersistedArtifact{}, err
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return PersistedArtifact{}, fmt.Errorf("create artifact dir: %w", err)
	}

	base := filepath.Base(src)
	targetName := fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), sanitizeFileName(base))
	targetPath := filepath.Join(targetDir, targetName)

	if err := copyFile(src, targetPath); err != nil {
		return PersistedArtifact{}, err
	}

	sum, err := checksum(targetPath)
	if err != nil {
		return PersistedArtifact{}, err
	}

	rel, err := filepath.Rel(r.Paths.Home, targetPath)
	if err != nil {
		return PersistedArtifact{}, fmt.Errorf("relative artifact path: %w", err)
	}

	return PersistedArtifact{
		RelativePath: filepath.ToSlash(rel),
		AbsolutePath: targetPath,
		OriginalName: base,
		SHA256:       sum,
	}, nil
}

func artifactDirForKind(paths Paths, kind string, ownerID string) (string, error) {
	switch kind {
	case "audio":
		return filepath.Join(paths.AudioDir, ownerID), nil
	case "photo":
		return filepath.Join(paths.PhotosDir, ownerID), nil
	case "document":
		return filepath.Join(paths.DocumentsDir, ownerID), nil
	default:
		return "", fmt.Errorf("unsupported artifact kind: %s", kind)
	}
}

func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", " ", "-")
	return replacer.Replace(name)
}

func copyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create target file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	if err := out.Sync(); err != nil {
		return fmt.Errorf("sync file: %w", err)
	}

	return nil
}

func checksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open checksum file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("hash file: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (r *Runtime) CreateBackup(ctx context.Context) (string, BackupManifest, error) {
	schemaVersion, err := r.Store.SchemaVersion(ctx)
	if err != nil {
		return "", BackupManifest{}, err
	}

	files, err := listFiles(r.Paths.ArtifactsDir)
	if err != nil {
		return "", BackupManifest{}, err
	}

	manifest := BackupManifest{
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		SchemaVersion: schemaVersion,
		DatabasePath:  "state/york.db",
		ArtifactRoot:  "artifacts",
		ArtifactCount: len(files),
	}

	backupPath := filepath.Join(r.Paths.BackupsDir, fmt.Sprintf("york-backup-%d.tar.gz", time.Now().UTC().Unix()))
	file, err := os.Create(backupPath)
	if err != nil {
		return "", BackupManifest{}, fmt.Errorf("create backup file: %w", err)
	}
	defer file.Close()

	gz := gzip.NewWriter(file)
	defer gz.Close()

	tw := tar.NewWriter(gz)
	defer tw.Close()

	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", BackupManifest{}, fmt.Errorf("marshal manifest: %w", err)
	}
	manifestBytes = append(manifestBytes, '\n')
	if err := writeTarEntry(tw, "manifest.json", manifestBytes, 0o644); err != nil {
		return "", BackupManifest{}, err
	}

	if err := addFileToTar(tw, r.Paths.Home, r.Paths.DBPath); err != nil {
		return "", BackupManifest{}, err
	}

	for _, path := range files {
		if err := addFileToTar(tw, r.Paths.Home, path); err != nil {
			return "", BackupManifest{}, err
		}
	}

	return backupPath, manifest, nil
}

func VerifyBackup(backupPath string) (BackupManifest, []string, error) {
	file, err := os.Open(backupPath)
	if err != nil {
		return BackupManifest{}, nil, fmt.Errorf("open backup: %w", err)
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return BackupManifest{}, nil, fmt.Errorf("open gzip backup: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	files := []string{}
	var manifest BackupManifest

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return BackupManifest{}, nil, fmt.Errorf("read backup entry: %w", err)
		}

		files = append(files, header.Name)
		if header.Name == "manifest.json" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return BackupManifest{}, nil, fmt.Errorf("read manifest entry: %w", err)
			}
			if err := json.Unmarshal(data, &manifest); err != nil {
				return BackupManifest{}, nil, fmt.Errorf("unmarshal manifest: %w", err)
			}
		}
	}

	if manifest.DatabasePath == "" {
		return BackupManifest{}, nil, fmt.Errorf("backup manifest missing database path")
	}

	return manifest, files, nil
}

func listFiles(root string) ([]string, error) {
	paths := []string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk files: %w", err)
	}
	return paths, nil
}

func addFileToTar(tw *tar.Writer, baseRoot string, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat backup file: %w", err)
	}
	if info.IsDir() {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open backup file: %w", err)
	}
	defer file.Close()

	rel, err := filepath.Rel(baseRoot, path)
	if err != nil {
		return fmt.Errorf("relative backup path: %w", err)
	}
	header := &tar.Header{
		Name:    filepath.ToSlash(rel),
		Mode:    0o644,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("write tar header: %w", err)
	}
	if _, err := io.Copy(tw, file); err != nil {
		return fmt.Errorf("write tar file: %w", err)
	}
	return nil
}

func writeTarEntry(tw *tar.Writer, name string, contents []byte, mode int64) error {
	header := &tar.Header{
		Name:    name,
		Mode:    mode,
		Size:    int64(len(contents)),
		ModTime: time.Now().UTC(),
	}
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("write tar manifest header: %w", err)
	}
	if _, err := tw.Write(contents); err != nil {
		return fmt.Errorf("write tar manifest: %w", err)
	}
	return nil
}
