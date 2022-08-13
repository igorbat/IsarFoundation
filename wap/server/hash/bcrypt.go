package hash // import "fastbot/server/hash"
import (
	"os/exec"
)

func WesnothBcrypt (salt, password string) (string, error){
	out, err := exec.Command("./wesnoth_bcrypt_auth", salt, password).Output()
    	if err != nil {
        	return "", err
    	}
	return string(out), nil
}
