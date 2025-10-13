package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

var (
	serverURL string
	alias     string
	ttlDays   int
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tinyurl",
		Short: "TinyURL CLI - сокращайте ссылки из командной строки",
	}

	rootCmd.PersistentFlags().StringVarP(&serverURL, "server", "s", "http://localhost:8080", "Адрес сервера TinyURL")

	shortCmd := &cobra.Command{
		Use:   "short [url]",
		Short: "Сократить URL",
		Args:  cobra.ExactArgs(1),
		RunE:  shortURL,
	}
	shortCmd.Flags().StringVarP(&alias, "alias", "a", "", "Пользовательский алиас для ссылки")
	shortCmd.Flags().IntVarP(&ttlDays, "ttl", "t", 0, "Срок жизни ссылки в днях (0 = бессрочно)")

	statsCmd := &cobra.Command{
		Use:   "stats [code]",
		Short: "Получить статистику по коду",
		Args:  cobra.ExactArgs(1),
		RunE:  getStats,
	}

	delCmd := &cobra.Command{
		Use:   "delete [code]",
		Short: "Удалить короткую ссылку",
		Args:  cobra.ExactArgs(1),
		RunE:  deleteURL,
	}
	rootCmd.AddCommand(delCmd)

	rootCmd.AddCommand(shortCmd, statsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func apiURL(base string, p string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(strings.TrimRight(u.Path, "/"), p)
	return u.String(), nil
}

func deleteURL(cmd *cobra.Command, args []string) error {
	code := args[0]
	u, err := apiURL(serverURL, "/delete/"+code)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent:
		fmt.Println("Deleted successfully")
		return nil
	case http.StatusNotFound:
		return fmt.Errorf("code %s not found", code)
	default:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}

func shortURL(cmd *cobra.Command, args []string) error {
	longURL := args[0]
	reqBody, err := json.Marshal(map[string]interface{}{
		"url":      longURL,
		"alias":    alias,
		"ttl_days": ttlDays,
	})
	if err != nil {
		return err
	}

	u, err := apiURL(serverURL, "/shorten")
	if err != nil {
		return err
	}

	resp, err := http.Post(u, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	switch code := resp.StatusCode; {
	case code >= 200 && code < 300:

		var result struct {
			Code     string `json:"code"`
			ShortURL string `json:"short_url"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("не удалось разобрать ответ: %w", err)
		}

		fmt.Println("Короткая ссылка:", result.ShortURL)
		return nil

	case code >= 300 && code < 400:
		loc := resp.Header.Get("Location")
		return fmt.Errorf("unexpected redirect %d → %s", code, loc)

	case code >= 400 && code < 500:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("client error %d: %s", code, strings.TrimSpace(string(body)))

	default:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error %d: %s", code, strings.TrimSpace(string(body)))
	}
}

func getStats(cmd *cobra.Command, args []string) error {
	code := args[0]

	u, err := apiURL(serverURL, "/stats/"+code)
	if err != nil {
		return err
	}

	resp, err := http.Get(u)
	if err != nil {
		return fmt.Errorf("ошибка при отправке GET-запроса к %s: %w", u, err)
	}
	defer resp.Body.Close()

	switch status := resp.StatusCode; {
	case status >= 200 && status < 300:
		var stats struct {
			URL       string  `json:"url"`
			CreatedAt string  `json:"created_at"`
			ExpiresAt *string `json:"expires_at,omitempty"`
			HitCount  int     `json:"hit_count"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return fmt.Errorf("не удалось разобрать ответ: %w", err)
		}

		fmt.Println("URL:", stats.URL)
		fmt.Println("Создано:", stats.CreatedAt)
		if stats.ExpiresAt != nil {
			fmt.Println("Истекает:", *stats.ExpiresAt)
		} else {
			fmt.Println("Истекает: никогда")
		}
		fmt.Println("Количество переходов:", stats.HitCount)
		return nil

	case status >= 300 && status < 400:
		loc := resp.Header.Get("Location")
		return fmt.Errorf("unexpected redirect %d → %s", status, loc)

	case status >= 400 && status < 500:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("client error %d: %s", status, strings.TrimSpace(string(body)))

	default:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error %d: %s", status, strings.TrimSpace(string(body)))
	}
}
