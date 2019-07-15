package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/encode", func(w http.ResponseWriter, r *http.Request) {
		resErr := func(status int) {
			w.WriteHeader(status)
			io.WriteString(w, http.StatusText(status))
		}
		if r.Method != http.MethodPost {
			resErr(http.StatusMethodNotAllowed)
			return
		}
		fmt.Println("POST", time.Now())

		err := r.ParseMultipartForm(10 * 1000 * 1000)
		if err != nil {
			fmt.Println(err)
			resErr(http.StatusBadRequest)
			return
		}

		images := r.MultipartForm.File["images"]
		num := len(images)
		if num == 0 {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "画像は1枚以上必要です")
			return
		}

		seconds, err := strconv.ParseFloat(r.FormValue("seconds"), 64)
		if err != nil {
			fmt.Println(err)
			resErr(http.StatusBadRequest)
			return
		}
		if seconds <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "秒数を指定して")
			return
		}

		tmpdir := os.TempDir()

		for i, img := range images {
			err := func() error {
				f, err := img.Open()
				if err != nil {
					return err
				}
				defer f.Close()
				f2, err := os.Create(filepath.Join(tmpdir, fmt.Sprintf("img_%03d.png", i+1)))
				if err != nil {
					return err
				}
				defer f2.Close()
				if _, err := io.Copy(f2, f); err != nil {
					return err
				}
				return nil
			}()
			if err != nil {
				fmt.Println(err)
				resErr(http.StatusInternalServerError)
				return
			}
			defer func(i int) {
				err = os.Remove(filepath.Join(tmpdir, fmt.Sprintf("img_%03d.png", i+1)))
				if err != nil {
					fmt.Println("remove", err)
					resErr(http.StatusInternalServerError)
					return
				}
			}(i)
		}

		wd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			resErr(http.StatusInternalServerError)
			return
		}
		out := filepath.Join(wd, "output.mp4")
		cmd := exec.Command("ffmpeg", []string{
			"-r", fmt.Sprintf("%d/%f", num, seconds),
			"-f", "image2",
			"-s", "1040x780",
			"-i", filepath.Join(tmpdir, "img_%03d.png"),
			"-vcodec", "libx264",
			"-pix_fmt", "yuv420p",
			"-y",
			out,
		}...)
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err = cmd.Run()
		if err != nil {
			fmt.Println(err)
			fmt.Println(buf.String())
			resErr(http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/result", http.StatusFound)
	})
	http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		out := filepath.Join(wd, "output.mp4")
		http.ServeFile(w, r, out)
	})
	http.ListenAndServe("0.0.0.0:1234", nil)
}
