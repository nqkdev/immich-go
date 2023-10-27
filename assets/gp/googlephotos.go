package gp

import (
	"context"
	"fmt"
	"immich-go/assets"
	"immich-go/helpers/fshelper"
	"io/fs"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Takeout struct {
	fsys        fs.FS
	filesByDir  map[string][]fileKey        // files name mapped by dir
	jsonByYear  map[jsonKey]*googleMetaData // JSON by year of capture and full path
	albumsByDir map[string]string           // album title mapped by dir
}

type fileKey struct {
	name string
	size int64
}

type jsonKey struct {
	year int
	name string
}
type Album struct {
	Title string
}

func NewTakeout(ctx context.Context, fsys fs.FS) (*Takeout, error) {
	to := Takeout{
		fsys:        fsys,
		filesByDir:  map[string][]fileKey{},
		jsonByYear:  map[jsonKey]*googleMetaData{},
		albumsByDir: map[string]string{},
	}
	err := to.walk(ctx, fsys)

	return &to, err
}

// walk the given FS to collect images file names and metadata files
func (to *Takeout) walk(ctx context.Context, fsys fs.FS) error {
	err := fs.WalkDir(fsys, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			// Check if the context has been cancelled
			return ctx.Err()
		default:
		}
		dir, base := path.Split(name)
		dir = strings.TrimSuffix(dir, "/")
		if dir == "" {
			dir = "."
		}
		if d.IsDir() {
			if base == "Failed Videos" {
				return fs.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(path.Ext(name))
		switch ext {
		case ".json":
			md, err := fshelper.ReadJSON[googleMetaData](fsys, name)
			if err == nil {
				if md.isAlbum() {
					to.albumsByDir[dir] = path.Base(dir)
				} else {
					key := jsonKey{
						year: md.PhotoTakenTime.Time().Year(),
						name: base,
					}
					if prevMD, exists := to.jsonByYear[key]; exists {
						// if prevMD.PhotoTakenTime != md.PhotoTakenTime {
						// 	fmt.Println("!surprise! 1+ json with different date", base)
						// }
						// if prevMD.Title != md.Title {
						// 	fmt.Println("!surprise! 1+ json with different title", base)
						// }
						prevMD.foundInPaths = append(prevMD.foundInPaths, dir)
						to.jsonByYear[key] = prevMD
					} else {
						md.foundInPaths = append(md.foundInPaths, dir)
						to.jsonByYear[key] = md
					}
				}
			}
		default:
			info, err := d.Info()
			if err != nil {
				return err
			}
			key := fileKey{name: base, size: info.Size()}
			l := to.filesByDir[dir]
			l = append(l, key)
			to.filesByDir[dir] = l
		}
		return nil
	})

	return err
}

// Browse gives back to the main program the list of assets with resolution of file name, album, dates...
func (to *Takeout) Browse(ctx context.Context) chan *assets.LocalAssetFile {
	c := make(chan *assets.LocalAssetFile)
	passed := map[fileKey]any{}
	go func() {
		defer close(c)
		for k, md := range to.jsonByYear {
			assets := to.jsonAssets(k, md)
			for _, a := range assets {
				fk := fileKey{name: path.Base(a.FileName), size: int64(a.FileSize)}
				if _, exist := passed[fk]; !exist {
					passed[fk] = nil
					select {
					case <-ctx.Done():
						return
					default:
						c <- a
					}
				}
			}
		}
	}()

	return c

}

// jsonAssets search assets that are linked to this JSON
//
//   the asset is named after the JSON name
//   the asset name can be 1 char shorter than the JSON name
//   but several assets can match with the JSON 🤯
//   the asset can be placed in another folder than the JSON
//   when the JSON is found in an album dir, the asset belongs to the album
//		but the image can be found in year's folder 🤯
//   the asset name is the JSON title field

func (to *Takeout) jsonAssets(key jsonKey, md *googleMetaData) []*assets.LocalAssetFile {

	var list []*assets.LocalAssetFile

	yearDir := path.Join(path.Dir(md.foundInPaths[0]), fmt.Sprintf("Photos from %d", md.PhotoTakenTime.Time().Year()))

	jsonInYear := false
	paths := md.foundInPaths
	for _, d := range md.foundInPaths {
		if d == yearDir {
			jsonInYear = true
			break
		}
	}
	if !jsonInYear {
		paths = append(paths, yearDir)
	}

	// Search for the assets in folders where the JSON has been found
	for _, d := range paths {
		l := to.filesByDir[d]

		for _, f := range l {

			matched := normalMatch(key.name, f.name)
			matched = matched || matchWithOneCharOmitted(key.name, f.name)
			matched = matched || matchVeryLongNameWithNumber(key.name, f.name)
			matched = matched || matchDuplicateInYear(key.name, f.name)
			matched = matched || matchEditedName(key.name, f.name)

			if matched {
				list = append(list, to.copyGoogleMDToAsset(md, path.Join(d, f.name), int(f.size)))
			}
		}
	}
	return list
}

// normalMatch
//
//	PXL_20230922_144936660.jpg.json
//	PXL_20230922_144936660.jpg
//
//	05yqt21kruxwwlhhgrwrdyb6chhwszi9bqmzu16w0 2.jp.json
//	05yqt21kruxwwlhhgrwrdyb6chhwszi9bqmzu16w0 2.jpg
func normalMatch(jsonName string, fileName string) bool {
	base := strings.TrimSuffix(jsonName, path.Ext(jsonName))
	ext := path.Ext(base)
	if ext == ".jp" {
		base += "g"
	}

	return base == fileName
}

// matchWithOneCharOmitted
//
//	PXL_20230809_203449253.LONG_EXPOSURE-02.ORIGIN.json
//	PXL_20230809_203449253.LONG_EXPOSURE-02.ORIGINA.jpg
func matchWithOneCharOmitted(jsonName string, fileName string) bool {
	base := strings.TrimSuffix(jsonName, path.Ext(jsonName))
	ext := path.Ext(base)
	switch ext {
	case "":
	default:
		if _, err := fshelper.MimeFromExt(ext); err == nil {
			return false
		}
	}
	fileName = strings.TrimSuffix(fileName, path.Ext(fileName))
	_, s := utf8.DecodeLastRuneInString(fileName)
	fileName = fileName[:len(fileName)-s]
	return base == fileName
}

// matchVeryLongNameWithNumber
//
//	Backyard_ceremony_wedding_photography_xxxxxxx_(494).json
//	Backyard_ceremony_wedding_photography_xxxxxxx_m(494).jpg
func matchVeryLongNameWithNumber(jsonName string, fileName string) bool {
	jsonName = strings.TrimSuffix(jsonName, path.Ext(jsonName))

	p1JSON := strings.Index(jsonName, "(")
	if p1JSON < 0 {
		return false
	}
	p2JSON := strings.Index(jsonName, ")")
	if p2JSON < 0 || p2JSON != len(jsonName)-1 {
		return false
	}
	p1File := strings.Index(fileName, "(")
	if p1File < 0 || p1File != p1JSON+1 {
		return false
	}
	if jsonName[:p1JSON] != fileName[:p1JSON] {
		return false
	}
	p2File := strings.Index(fileName, ")")
	return jsonName[p1JSON+1:p2JSON] == fileName[p1File+1:p2File]
}

// matchDuplicateInYear
//
//	IMG_3479.JPG(2).json
//	IMG_3479(2).JPG
func matchDuplicateInYear(jsonName string, fileName string) bool {
	jsonName = strings.TrimSuffix(jsonName, path.Ext(jsonName))
	p1JSON := strings.Index(jsonName, "(")
	if p1JSON < 1 {
		return false
	}
	p2JSON := strings.Index(jsonName, ")")
	if p2JSON < 0 || p2JSON != len(jsonName)-1 {
		return false
	}

	num := jsonName[p1JSON:]
	jsonName = strings.TrimSuffix(jsonName, num)
	ext := path.Ext(jsonName)
	jsonName = strings.TrimSuffix(jsonName, ext) + num + ext
	return jsonName == fileName
}

// matchEditedName
//   PXL_20220405_090123740.PORTRAIT.jpg.json
//   PXL_20220405_090123740.PORTRAIT.jpg
//   PXL_20220405_090123740.PORTRAIT-modifié.jpg

func matchEditedName(jsonName string, fileName string) bool {
	base := strings.TrimSuffix(jsonName, path.Ext(jsonName))
	ext := path.Ext(base)
	if ext != "" {
		if _, err := fshelper.MimeFromExt(ext); err == nil {
			base := strings.TrimSuffix(base, ext)
			fileName = strings.TrimSuffix(fileName, path.Ext(fileName))
			return strings.HasPrefix(fileName, base)
		}
	}
	return false
}

func (to *Takeout) copyGoogleMDToAsset(md *googleMetaData, filename string, length int) *assets.LocalAssetFile {
	a := assets.LocalAssetFile{
		FileName:    filename,
		FileSize:    length,
		Title:       md.Title,
		Altitude:    md.GeoDataExif.Altitude,
		Latitude:    md.GeoDataExif.Latitude,
		Longitude:   md.GeoDataExif.Longitude,
		Archived:    md.Archived,
		FromPartner: md.isPartner(),
		Trashed:     md.Trashed,
		DateTaken:   md.PhotoTakenTime.Time(),
		FSys:        to.fsys,
	}
	for _, p := range md.foundInPaths {
		if album, exists := to.albumsByDir[p]; exists {
			a.Albums = append(a.Albums, album)
		}

	}
	return &a
}

var (
	numberedName         = regexp.MustCompile(`(?m)(.*)(\..+)(\(\d+\))(\.\w+)$`)
	veryLongNumberedName = regexp.MustCompile(`(?m)(.+)(.)(\(\.+\))(\.\w+$)`)
)
