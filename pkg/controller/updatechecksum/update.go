package updatechecksum

import (
	"context"
	"fmt"
	"io"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	finder "github.com/aquaproj/aqua/pkg/config-finder"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Controller struct {
	rootDir           string
	configFinder      ConfigFinder
	configReader      domain.ConfigReader
	registryInstaller domain.RegistryInstaller
	fs                afero.Fs
	runtime           *runtime.Runtime
	chkDL             domain.ChecksumDownloader
}

type ConfigFinder interface {
	Finds(wd, configFilePath string) []string
}

func New(param *config.Param, configFinder ConfigFinder, configReader domain.ConfigReader, registInstaller domain.RegistryInstaller, fs afero.Fs, rt *runtime.Runtime, chkDL domain.ChecksumDownloader) *Controller {
	return &Controller{
		rootDir:           param.RootDir,
		configFinder:      configFinder,
		configReader:      configReader,
		registryInstaller: registInstaller,
		fs:                fs,
		runtime:           rt,
		chkDL:             chkDL,
	}
}

func (ctrl *Controller) UpdateChecksum(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	for _, cfgFilePath := range ctrl.configFinder.Finds(param.PWD, param.ConfigFilePath) {
		if err := ctrl.updateChecksum(ctx, logE, cfgFilePath); err != nil {
			return err
		}
	}

	return ctrl.updateChecksumAll(ctx, logE, param)
}

func (ctrl *Controller) updateChecksumAll(ctx context.Context, logE *logrus.Entry, param *config.Param) error {
	if !param.All {
		return nil
	}
	for _, cfgFilePath := range param.GlobalConfigFilePaths {
		if _, err := ctrl.fs.Stat(cfgFilePath); err != nil {
			continue
		}
		if err := ctrl.updateChecksum(ctx, logE, cfgFilePath); err != nil {
			return err
		}
	}
	return nil
}

func (ctrl *Controller) updateChecksum(ctx context.Context, logE *logrus.Entry, cfgFilePath string) error {
	cfg := &aqua.Config{}
	if cfgFilePath == "" {
		return finder.ErrConfigFileNotFound
	}
	if err := ctrl.configReader.Read(cfgFilePath, cfg); err != nil {
		return err //nolint:wrapcheck
	}

	registryContents, err := ctrl.registryInstaller.InstallRegistries(ctx, cfg, cfgFilePath, logE)
	if err != nil {
		return err //nolint:wrapcheck
	}

	checksums := checksum.New()
	checksumFilePath := checksum.GetChecksumFilePathFromConfigFilePath(cfgFilePath)
	if err := checksums.ReadFile(ctrl.fs, checksumFilePath); err != nil {
		return fmt.Errorf("read a checksum JSON: %w", err)
	}
	pkgs, _ := config.ListPackages(logE, cfg, ctrl.runtime, registryContents)
	parser := &checksum.FileParser{}
	for _, pkg := range pkgs {
		if err := ctrl.updatePackage(ctx, logE, checksums, parser, pkg); err != nil {
			return err
		}
	}
	if err := checksums.UpdateFile(ctrl.fs, checksumFilePath); err != nil {
		return fmt.Errorf("update a checksum file: %w", err)
	}
	return nil
}

func (ctrl *Controller) updatePackage(ctx context.Context, logE *logrus.Entry, checksums *checksum.Checksums, parser *checksum.FileParser, pkg *config.Package) error {
	if pkg.PackageInfo.Checksum == nil {
		logE.WithFields(logrus.Fields{
			"package_name":    pkg.Package.Name,
			"package_version": pkg.Package.Version,
		}).Debug("the package doesn't support getting checksum file")
		return nil
	}
	logE.Info("downloading a checksum file")
	file, _, err := ctrl.chkDL.DownloadChecksum(ctx, logE, pkg)
	if err != nil {
		return fmt.Errorf("download a checksum file: %w", err)
	}
	if file == nil {
		logE.Info("checksum file isn't found")
		return nil
	}
	defer file.Close()
	b, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read a checksum file: %w", err)
	}
	m, err := parser.ParseChecksumFile(string(b), pkg)
	if err != nil {
		return fmt.Errorf("parse a checksum file: %w", err)
	}
	for asset, chksum := range m {
		chkID, err := pkg.GetChecksumIDFromAsset(ctrl.runtime, asset)
		if err != nil {
			return fmt.Errorf("get checksum ID from asset: %w", err)
		}
		logE.WithFields(logrus.Fields{
			"checksum_id": chkID,
			"checksum":    chksum,
			"asset":       asset,
		}).Debug("set a checksum")
		checksums.Set(chkID, chksum)
	}
	return nil
}
