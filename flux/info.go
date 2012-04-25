package flux

import (
	"go/build"
	."os"
	."fmt"
	"path"
	"path/filepath"
)

var packageInfo chan *PackageInfo = make(chan *PackageInfo)

func getPackageInfo(pathStr string) *PackageInfo {
	packageInfo := &PackageInfo{filepath.Base(pathStr), "", []*PackageInfo{}, nil}
	
	pkg, err := build.ImportDir(pathStr, build.FindOnly & build.AllowBinary)
	if err == nil && !pkg.IsCommand() {
		packageInfo.ImportPath = pkg.ImportPath
	} else if err != nil {
		if _, ok := err.(*build.NoGoError); !ok { Println(err) }
	}
	
	if file, err := Open(pathStr); err == nil {
		if fileInfos, err := file.Readdir(-1); err == nil {
			for _, fileInfo := range fileInfos {
				if fileInfo.IsDir() {
					subPackageInfo := getPackageInfo(filepath.Join(pathStr, fileInfo.Name()))
					if len(subPackageInfo.SubPackages) > 0 || len(subPackageInfo.ImportPath) > 0 {
						subPackageInfo.Parent = packageInfo
						packageInfo.SubPackages = append(packageInfo.SubPackages, subPackageInfo)
					}
				}
			}
		}
	}
	
	if len(packageInfo.ImportPath) == 0 && len(packageInfo.SubPackages) == 1 && len(packageInfo.SubPackages[0].ImportPath) == 0 {
		subPackageInfo := packageInfo.SubPackages[0]
		subPackageInfo.Name = path.Join(packageInfo.Name, subPackageInfo.Name)
		packageInfo = subPackageInfo
	}
	
	return packageInfo
}

func init() {
	go func() {
		rootPackageInfo := &PackageInfo{"", "", []*PackageInfo{}, nil}
		for _, srcDir := range build.Default.SrcDirs() {
			subPackageInfo := getPackageInfo(srcDir)
			for _, p := range subPackageInfo.SubPackages {
				p.Parent = rootPackageInfo
			}
			rootPackageInfo.SubPackages = append(rootPackageInfo.SubPackages, subPackageInfo.SubPackages...)
		}
		for { packageInfo <- rootPackageInfo }
	}()
}

type PackageInfo struct {
	Name string
	ImportPath string
	SubPackages []*PackageInfo
	Parent *PackageInfo
}

func GetPackageInfo() *PackageInfo {
	return <-packageInfo
}
