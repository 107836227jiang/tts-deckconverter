package mtg

import (
	"net/url"
	"path"
	"regexp"
	"strings"

	scryfall "github.com/BlueMonday/go-scryfall"

	"github.com/jeandeaual/tts-deckconverter/plugins"
)

type imageQuality string

const (
	small  imageQuality = "small"
	normal imageQuality = "normal"
	large  imageQuality = "large"
	png    imageQuality = "png"
)

type magicPlugin struct {
	id   string
	name string
}

func (p magicPlugin) PluginID() string {
	return p.id
}

func (p magicPlugin) PluginName() string {
	return p.name
}

func (p magicPlugin) SupportedLanguages() []string {
	return []string{
		string(scryfall.LangEnglish),
		string(scryfall.LangSpanish),
		string(scryfall.LangFrench),
		string(scryfall.LangGerman),
		string(scryfall.LangItalian),
		string(scryfall.LangPortuguese),
		string(scryfall.LangJapanese),
		string(scryfall.LangKorean),
		string(scryfall.LangRussian),
		string(scryfall.LangSimplifiedChinese),
		string(scryfall.LangTraditionalChinese),
	}
}

func (p magicPlugin) AvailableOptions() plugins.Options {
	return plugins.Options{
		"quality": plugins.Option{
			Type:        plugins.OptionTypeEnum,
			Description: "image quality",
			AllowedValues: []string{
				string(small),
				string(normal),
				string(large),
				string(png),
			},
			DefaultValue: string(large),
		},
		"rulings": plugins.Option{
			Type:         plugins.OptionTypeBool,
			Description:  "add the rulings to each card description",
			DefaultValue: false,
		},
	}
}

func (p magicPlugin) URLHandlers() []plugins.URLHandler {
	return []plugins.URLHandler{
		{
			BasePath: "https://scryfall.com",
			Regex:    regexp.MustCompile(`^https://scryfall\.com/@.+/decks/`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				parsedURL, err := url.Parse(baseURL)
				if err != nil {
					return nil, err
				}

				uuid := path.Base(parsedURL.Path)

				return handleLink(
					baseURL,
					`//h1[contains(@class,'deck-details-title')]`,
					"https://api.scryfall.com/decks/"+uuid+"/export/text",
					options,
				)
			},
		},
		{
			BasePath: "https://deckstats.net",
			Regex:    regexp.MustCompile(`^https://deckstats\.net/decks/`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				fileURL, err := url.Parse(baseURL)
				if err != nil {
					return nil, err
				}
				q := fileURL.Query()
				q.Set("include_comments", "1")
				q.Set("export_mtgarena", "1")
				fileURL.RawQuery = q.Encode()

				return handleLink(
					baseURL,
					`//h2[@id='subtitle']`,
					fileURL.String(),
					options,
				)
			},
		},
		{
			BasePath: "https://tappedout.net",
			Regex:    regexp.MustCompile(`^https://tappedout\.net/mtg-decks/`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				fileURL, err := url.Parse(baseURL)
				if err != nil {
					return nil, err
				}
				q := fileURL.Query()
				q.Set("fmt", "dec")
				fileURL.RawQuery = q.Encode()

				return handleLink(
					baseURL,
					`//div[contains(@class,'well')]/h2/text()`,
					fileURL.String(),
					options,
				)
			},
		},
		{
			BasePath: "https://deckbox.org",
			Regex:    regexp.MustCompile(`^https://deckbox\.org/sets/`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				var fileURL string

				if strings.HasSuffix(baseURL, "/") {
					fileURL = baseURL + "export"
				} else {
					fileURL = baseURL + "/export"
				}

				return handleHTMLLink(
					baseURL,
					`//span[@id='deck_built_title']/following-sibling::text()`,
					fileURL,
					options,
				)
			},
		},
		{
			BasePath: "https://www.mtggoldfish.com",
			Regex:    regexp.MustCompile(`^https://www\.mtggoldfish\.com/deck/`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				return handleLinkWithDownloadLink(
					baseURL,
					`//h1[contains(@class,'deck-view-title')]/text()`,
					`//a[contains(text(),'Download')]/@href`,
					"https://www.mtggoldfish.com",
					options,
				)
			},
		},
		{
			BasePath: "https://manastack.com",
			Regex:    regexp.MustCompile(`^https://manastack\.com/deck/`),
			Handler:  handleManaStackLink,
		},
	}
}

func (p magicPlugin) FileExtHandlers() map[string]plugins.FileHandler {
	return map[string]plugins.FileHandler{
		".dec": fromDeckFile,
		".cod": fromCockatriceDeckFile,
	}
}

func (p magicPlugin) GenericFileHandler() plugins.PathHandler {
	return parseFile
}

func (p magicPlugin) AvailableBacks() map[string]plugins.Back {
	return map[string]plugins.Back{
		plugins.DefaultBackKey: {
			URL:         defaultBackURL,
			Description: "standard paper card back",
		},
		"planechase": {
			URL:         planechaseBackURL,
			Description: "Planechase Plane / Phenomenon card back",
		},
		"archenemy": {
			URL:         archenemyBackURL,
			Description: "Archenemy Scheme card back",
		},
		"m_filler": {
			URL:         mFillerBackURL,
			Description: "filler card back with a white M in the middle",
		},
	}
}

// MagicPlugin is the exported plugin for this package
var MagicPlugin = magicPlugin{
	id:   "mtg",
	name: "Magic",
}
