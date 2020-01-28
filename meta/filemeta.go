package meta

import (
	"GoDrive/db"
	"sort"
)

// FileMeta contains file meta info struct
type FileMeta struct {
	FileSha1 string `json:"hashkey"`
	FileName string `json:"name"`
	FileSize int64  `json:"size"`
	Location string `json:"location"`
	UploadAt string `json:"date"`
}

var fileMetas map[string]FileMeta

// when import this package, init() will be called
func init() {
	fileMetas = make(map[string]FileMeta)
}

// UpdateFileMeta : add/modify file meta info in RAM
func UpdateFileMeta(fm FileMeta) {
	fileMetas[fm.FileSha1] = fm
}

// UpdateFileMetaDB : add/modify file meta into DB
func UpdateFileMetaDB(fm FileMeta) bool {
	return db.OnFileUploadFinished(fm.FileSha1, fm.FileName, fm.FileSize, fm.Location)
}

// GetFileMeta : get FileMeta struct based on give SHA1 hash code
func GetFileMeta(sha1 string) FileMeta {
	return fileMetas[sha1]
}

// GetFileMetaDB : get file meta info from DB
func GetFileMetaDB(sha1 string) (FileMeta, error) {
	tFile, err := db.GetFileMeta(sha1)
	if err != nil {
		return FileMeta{}, err
	}

	fMeta := FileMeta{
		FileSha1: tFile.FileHash,
		FileName: tFile.FileName.String,
		FileSize: tFile.FileSize.Int64,
		Location: tFile.FileLocation.String,
	}

	return fMeta, nil
}

// GetLastFileMetas : get the last `count` files' meta datas
func GetLastFileMetas(count int) []FileMeta {
	count = minInt(count, len(fileMetas))
	fMetaSlice := make([]FileMeta, len(fileMetas))
	for _, v := range fileMetas {
		fMetaSlice = append(fMetaSlice, v)
	}
	// sorted by 'uploadAt'
	sort.Sort(SortedByUploadTime(fMetaSlice))
	return fMetaSlice[0:count]
}

// GetLastFileMetasDB : get last `limit` files meta from DB
func GetLastFileMetasDB(limit int) ([]FileMeta, error) {
	files, err := db.GetLastNMetaList(limit)
	if err != nil {
		return make([]FileMeta, 0), err
	}

	fMetas := make([]FileMeta, len(files))
	for i := 0; i < len(fMetas); i++ {
		fMetas[i] = FileMeta{
			FileSha1: files[i].FileHash,
			FileName: files[i].FileName.String,
			FileSize: files[i].FileSize.Int64,
			Location: files[i].FileLocation.String,
		}
	}

	return fMetas, nil
}

// RemoveMeta : remove the file meta, in the future, need to consider about multithreading security
func RemoveMeta(fileSha1 string) {
	delete(fileMetas, fileSha1)
}

// RemoveMetaDB removes a file meta from the db
func RemoveMetaDB(filesha string) bool {
	return db.OnFileRemoved(filesha)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
