package cmdline

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

// CmdLine ...
type CmdLine struct {
	Command       string
	OutBuffer     string
	DisplayBuffer bool

	// Pour les requêtes avec des arguments dans lesquels il y a des espaces, il faut mettre "true"
	// Exemple : /rsync -ahv -e "ssh -p xxxx -i /home/user/.ssh/serverx" --rsync-path="sudo rsync" --delete ...
	MultiArgs bool

	// UseLog permet de créer un fichier log pour chaque utilisation de CmdLine.Run
	UseLog bool
}

// NewCmdLine ...
func NewCmdLine(display bool) *CmdLine {
	return &CmdLine{DisplayBuffer: display, MultiArgs: false}
}

// UseMultiArgs utilisé pour les requêtes avec des arguments dans lesquels il y a des espaces, il faut mettre "true"
// Exemple : /rsync -ahv -e "ssh -p xxxx -i /home/user/.ssh/serverx" --rsync-path="sudo rsync" --delete ...
func (c *CmdLine) UseMultiArgs() *CmdLine {
	c.MultiArgs = true
	return c
}

// Run permet d’exécuter des commandes système.
func (c *CmdLine) Run(command string) (string, error) {
	// Déclarer les variables
	var cmd *exec.Cmd
	var fileLog *os.File
	// Vider les retours précédents, ceci est important quand CmdLine est instancié une seule fois.
	c.OutBuffer = ""

	if c.MultiArgs {
		// Pour les requêtes avec des arguments dans lesquels il y a des espaces, il faut utiliser cette version.
		// Exemple : /rsync -ahv -e "ssh -p xxxx -i /home/user/.ssh/serverx" --rsync-path="sudo rsync" --delete ...
		cmd = exec.Command("sh", "-c", command)
	} else {
		cmdArgs := strings.Fields(command)
		cmd = exec.Command(cmdArgs[0], cmdArgs[1:len(cmdArgs)]...)
	}

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return c.OutBuffer, err
	}

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return c.OutBuffer, err
	}

	// execute async
	if err2 := cmd.Start(); err != nil {
		return c.OutBuffer, err2
	}

	// read stdout while executing
	in := bufio.NewScanner(stdout)

	// Log : Si le fichier n'existe pas, créez-le ou ajoutez-le au fichier
	if c.UseLog {
		fileLog, err = os.OpenFile(fmt.Sprintf("/tmp/mgo_cmd_%d.log", time.Now().UnixNano()), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalln(err)
		}
	}

	for in.Scan() {
		if c.DisplayBuffer {
			fmt.Println(in.Text())
		} else {
			c.OutBuffer += in.Text() + "\n"
		}

		// Log : Écrire la sortie dans le fichier log
		if c.UseLog {
			if _, err := fileLog.Write([]byte(in.Text() + "\n")); err != nil {
				log.Fatalln(err)
			}
		}
	}

	// Test : Fermer le fichier log
	if c.UseLog {
		if err := fileLog.Close(); err != nil {
			fmt.Println(err)
		}
	}

	if err := in.Err(); err != nil {
		return c.OutBuffer, err
	}

	// read stderr while executing
	inError := bufio.NewScanner(stderr)

	for inError.Scan() {
		if c.DisplayBuffer {
			fmt.Println(inError.Text())
		} else {
			c.OutBuffer += inError.Text() + "\n"
		}

		// Log : Écrire la sortie dans le fichier log
		if c.UseLog {
			if _, err := fileLog.Write([]byte(inError.Text() + "\n")); err != nil {
				log.Fatalln(err)
			}
		}
	}

	if err := inError.Err(); err != nil {
		return c.OutBuffer, err
	}

	cmd.Wait()

	return c.OutBuffer, nil
}
