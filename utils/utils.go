package utils

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// ReadLine renvoie le texte saisi par l'utilisateur dans la console.
func ReadLine() string {
	bio := bufio.NewReader(os.Stdin)
	line, _, err := bio.ReadLine()
	if err != nil {
		fmt.Println(err)
	}
	return string(line)
}

// GetMD5Hash génère une chaîne en MD5.
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// RunningTime voir : utils.Track
func RunningTime(s string) (string, time.Time) {
	log.Println("Start:	", s)
	return s, time.Now()
}

// Track permet de mesurer le temps d’exécuter d'un programme.
// Exemple d’utilisation dans d'une fonction : defer utils.Track(utils.RunningTime("CopyDir"))
func Track(s string, startTime time.Time) {
	endTime := time.Now()
	log.Println("End: ", s+".", "Total :", endTime.Sub(startTime))
}

// CopyFile copie le contenu du fichier nommé src dans le fichier nommé
// par dst. Le fichier sera créé s'il n'existe pas déjà. Si la
// le fichier de destination existe, tout son contenu sera remplacé par le contenu
// du fichier source. Le mode fichier sera copié à partir de la source et
// les données copiées sont synchronisées / vidées dans un stockage stable.
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}

	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	err = os.Chtimes(dst, si.ModTime(), si.ModTime())
	if err != nil {
		return
	}

	return
}

// CopyDir copie récursivement une arborescence de répertoires, en essayant de conserver les permissions.
// Le répertoire source doit exister, le répertoire de destination ne doit pas * exister *.
// Les liens symboliques sont ignorés et ignorés.
func CopyDir(src string, dst string, crush bool) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("Source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil && crush == false {
		return fmt.Errorf("Destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath, crush)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

// FileExists permet de savoir si un fichier existe vraiment.
func FileExists(name string) bool {
	if fi, err := os.Stat(name); err == nil {
		if fi.Mode().IsRegular() {
			return true
		}
	}
	return false
}

// DirExists permet de savoir si un dossier existe vraiment.
func DirExists(name string) bool {
	if fi, err := os.Stat(name); err == nil {
		if fi.Mode().IsDir() {
			return true
		}
	}
	return false
}
