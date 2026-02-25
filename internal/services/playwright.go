package services

import (
	"go_parser/internal/utils"

	"github.com/playwright-community/playwright-go"
)

// FetchPageWithPlaywright открывает страницу через Playwright и возвращает её HTML-содержимое.
func FetchPageWithPlaywright(url string) (string, error) {
	utils.Logger.Println("Запуск Playwright...")
	pw, err := playwright.Run()
	if err != nil {
		return "", err
	}
	defer pw.Stop()
	utils.Logger.Println("Playwright успешно запущен.")

	utils.Logger.Println("Запуск headless-браузера...")
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true), // Headless-режим
	})
	if err != nil {
		return "", err
	}
	defer browser.Close()
	utils.Logger.Println("Браузер успешно запущен.")

	utils.Logger.Println("Создание новой страницы...")
	page, err := browser.NewPage()
	if err != nil {
		return "", err
	}
	defer page.Close()
	utils.Logger.Println("Страница успешно создана.")

	utils.Logger.Printf("Переход по URL: %s\n", url)
	if _, err := page.Goto(url); err != nil {
		return "", err
	}
	utils.Logger.Println("Успешно перешли по URL.")

	// Ожидание загрузки страницы
	utils.Logger.Println("Ожидание загрузки страницы...")
	err = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	if err != nil {
		return "", err
	}
	utils.Logger.Println("Страница успешно загружена.")

	return page.Content()
}
