package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	geminiTTSAPIEndpoint     = "https://texttospeech.googleapis.com/v1/text:synthesize"
	defaultGeminiTTSModel    = "gemini-2.5-flash-tts"
	defaultGeminiTTSVoice    = "Callirrhoe"
	timeFormatForTTSFilename = "20060102-150405"
)

// hardcoded list of voices based on documentation
var availableGeminiVoices = []string{
	"Achernar",
	"Achird",
	"Algenib",
	"Algieba",
	"Alnilam",
	"Aoede",
	"Autonoe",
	"Callirrhoe",
	"Charon",
	"Despina",
	"Enceladus",
	"Erinome",
	"Fenrir",
	"Gacrux",
	"Iapetus",
	"Kore",
	"Laomedeia",
	"Leda",
	"Orus",
	"Pulcherrima",
	"Puck",
	"Rasalgethi",
	"Sadachbia",
	"Sadaltager",
	"Schedar",
	"Sulafat",
	"Umbriel",
	"Vindemiatrix",
	"Zephyr",
	"Zubenelgenubi",
}

// geminiLanguageCodeMap holds the supported languages.
var geminiLanguageCodeMap = map[string]string{
	"arabic (egypt)":              "ar-EG",
	"dutch (netherlands)":         "nl-NL",
	"english (india)":             "en-IN",
	"english (united states)":     "en-US",
	"french (france)":             "fr-FR",
	"german (germany)":            "de-DE",
	"hindi (india)":               "hi-IN",
	"indonesian (indonesia)":      "id-ID",
	"italian (italy)":             "it-IT",
	"japanese (japan)":            "ja-JP",
	"korean (south korea)":        "ko-KR",
	"marathi (india)":             "mr-IN",
	"polish (poland)":             "pl-PL",
	"portuguese (brazil)":         "pt-BR",
	"romanian (romania)":          "ro-RO",
	"russian (russia)":            "ru-RU",
	"spanish (spain)":             "es-ES",
	"tamil (india)":               "ta-IN",
	"telugu (india)":              "te-IN",
	"thai (thailand)":             "th-TH",
	"turkish (turkey)":            "tr-TR",
	"ukrainian (ukraine)":         "uk-UA",
	"vietnamese (vietnam)":        "vi-VN",
	"afrikaans (south africa)":    "af-ZA",
	"albanian (albania)":          "sq-AL",
	"amharic (ethiopia)":          "am-ET",
	"arabic (world)":              "ar-001",
	"armenian (armenia)":          "hy-AM",
	"azerbaijani (azerbaijan)":    "az-AZ",
	"bangla (bangladesh)":         "bn-BD",
	"basque (spain)":              "eu-ES",
	"belarusian (belarus)":        "be-BY",
	"bulgarian (bulgaria)":        "bg-BG",
	"burmese (myanmar)":           "my-MM",
	"catalan (spain)":             "ca-ES",
	"cebuano (philippines)":       "ceb-PH",
	"chinese, mandarin (china)":   "cmn-CN",
	"chinese, mandarin (taiwan)":  "cmn-TW",
	"croatian (croatia)":          "hr-HR",
	"czech (czech republic)":      "cs-CZ",
	"danish (denmark)":            "da-DK",
	"english (australia)":         "en-AU",
	"english (united kingdom)":    "en-GB",
	"estonian (estonia)":          "et-EE",
	"filipino (philippines)":      "fil-PH",
	"finnish (finland)":           "fi-FI",
	"french (canada)":             "fr-CA",
	"galician (spain)":            "gl-ES",
	"georgian (georgia)":          "ka-GE",
	"greek (greece)":              "el-GR",
	"gujarati (india)":            "gu-IN",
	"haitian creole (haiti)":      "ht-HT",
	"hebrew (israel)":             "he-IL",
	"hungarian (hungary)":         "hu-HU",
	"icelandic (iceland)":         "is-IS",
	"javanese (java)":             "jv-JV",
	"kannada (india)":             "kn-IN",
	"konkani (india)":             "kok-IN",
	"lao (laos)":                  "lo-LA",
	"latin (vatican city)":        "la-VA",
	"latvian (latvia)":            "lv-LV",
	"lithuanian (lithuania)":      "lt-IT",
	"luxembourgish (luxembourg)":  "lb-LU",
	"macedonian (north macedonia)": "mk-MK",
	"maithili (india)":            "mai-IN",
	"malagasy (madagascar)":       "mg-MG",
	"malay (malaysia)":            "ms-MY",
	"malayalam (india)":           "ml-IN",
	"mongolian (mongolia)":        "mn-MN",
	"nepali (nepal)":              "ne-NP",
	"norwegian, bokmål (norway)":  "nb-NO",
	"norwegian, nynorsk (norway)": "nn-NO",
	"odia (india)":                "or-IN",
	"pashto (afghanistan)":        "ps-AF",
	"persian (iran)":              "fa-IR",
	"portuguese (portugal)":       "pt-PT",
	"punjabi (india)":             "pa-IN",
	"serbian (serbia)":            "sr-RS",
	"sindhi (india)":              "sd-IN",
	"sinhala (sri lanka)":         "si-LK",
	"slovak (slovakia)":           "sk-SK",
	"slovenian (slovenia)":        "sl-SI",
	"spanish (latin america)":     "es-419",
	"spanish (mexico)":            "es-MX",
	"swahili (kenya)":             "sw-KE",
	"swedish (sweden)":            "sv-SE",
	"urdu (pakistan)":             "ur-PK",
}

// --- Resource Handler ---

func geminiLanguageCodesHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	jsonData, err := json.MarshalIndent(geminiLanguageCodeMap, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal language codes: %w", err)
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "gemini://language_codes",
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

// --- Tool Handlers ---

// listGeminiVoicesHandler handles the 'list_gemini_voices' tool request.
// It returns a hardcoded list of available Gemini TTS voices.
func listGeminiVoicesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Println("Handling list_gemini_voices request.")

	voiceListJSON, err := json.MarshalIndent(availableGeminiVoices, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal voice list: %v", err)), nil
	}

	summary := fmt.Sprintf("Found %d available Gemini TTS voices.", len(availableGeminiVoices))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: summary},
			mcp.TextContent{Type: "text", Text: string(voiceListJSON)},
		},
	}, nil
}

var audioEncodingToFileExtension = map[string]string{
	"LINEAR16": ".wav",
	"MP3":      ".mp3",
	"OGG_OPUS": ".ogg",
	"MULAW":    ".mulaw",
	"ALAW":     ".alaw",
	"PCM":      ".pcm",
	"M4A":      ".m4a",
}

var audioEncodingToMIMEType = map[string]string{
	"LINEAR16": "audio/wav",
	"MP3":      "audio/mpeg",
	"OGG_OPUS": "audio/ogg",
	"MULAW":    "audio/mulaw",
	"ALAW":     "audio/alaw",
	"PCM":      "audio/pcm",
	"M4A":      "audio/mp4",
}

// geminiAudioTTSHandler handles the 'gemini_audio_tts' tool request.
func geminiAudioTTSHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Printf("Handling gemini_audio_tts request with arguments: %v", request.GetArguments())

	// --- 1. Parse and Validate Arguments ---
	text, ok := request.GetArguments()["text"].(string)
	if !ok || strings.TrimSpace(text) == "" {
		return mcp.NewToolResultError("text parameter must be a non-empty string and is required"), nil
	}
	if len(text) > 800 {
		return mcp.NewToolResultError("text parameter cannot exceed 800 characters"), nil
	}

	prompt, _ := request.GetArguments()["prompt"].(string)

	modelName, _ := request.GetArguments()["model_name"].(string)
	if modelName == "" {
		modelName = defaultGeminiTTSModel
	}

	voiceName, _ := request.GetArguments()["voice_name"].(string)
	if voiceName == "" {
		voiceName = defaultGeminiTTSVoice
	}
	// Validate voice
	validVoice := false
	for _, v := range availableGeminiVoices {
		if v == voiceName {
			validVoice = true
			break
		}
	}
	if !validVoice {
		return mcp.NewToolResultError(fmt.Sprintf("invalid voice_name '%s'. Use 'list_gemini_voices' to see available voices", voiceName)), nil
	}

	languageCode, _ := request.GetArguments()["language_code"].(string)
	if languageCode == "" {
		languageCode = "en-US"
	}

	audioEncoding, _ := request.GetArguments()["audio_encoding"].(string)
	if audioEncoding == "" {
		audioEncoding = "LINEAR16"
	}

	outputDir, _ := request.GetArguments()["output_directory"].(string)
	filenamePrefix, _ := request.GetArguments()["output_filename_prefix"].(string)
	if filenamePrefix == "" {
		filenamePrefix = "gemini_tts_audio"
	}

	// --- 2. Call the TTS API ---
	audioBytes, err := callGeminiTTSAPIWithSDK(ctx, text, prompt, voiceName, modelName, audioEncoding, languageCode)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("error calling Gemini TTS API: %v", err)), nil
	}

	// --- 3. Process the Audio Response ---
	var contentItems []mcp.Content
	var fileSaveMessage string

	fileExtension, ok := audioEncodingToFileExtension[audioEncoding]
	if !ok {
		fileExtension = ".wav"
	}
	mimeType, ok := audioEncodingToMIMEType[audioEncoding]
	if !ok {
		mimeType = "audio/wav"
	}

	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fileSaveMessage = fmt.Sprintf("Error creating directory %s: %v. Audio data will be returned in response instead.", outputDir, err)
			log.Print(fileSaveMessage)
			// Fallback to returning data in response
			base64AudioData := base64.StdEncoding.EncodeToString(audioBytes)
			contentItems = append(contentItems, mcp.AudioContent{Type: "audio", Data: base64AudioData, MIMEType: mimeType})
		} else {
			filename := fmt.Sprintf("%s-%s-%s%s", filenamePrefix, voiceName, time.Now().Format(timeFormatForTTSFilename), fileExtension)
			savedFilename := filepath.Join(outputDir, filename)
			if err := os.WriteFile(savedFilename, audioBytes, 0644); err != nil {
				fileSaveMessage = fmt.Sprintf("Error writing audio file %s: %v. Audio data will be returned in response instead.", savedFilename, err)
				log.Print(fileSaveMessage)
				base64AudioData := base64.StdEncoding.EncodeToString(audioBytes)
				contentItems = append(contentItems, mcp.AudioContent{Type: "audio", Data: base64AudioData, MIMEType: mimeType})
			} else {
				fileSaveMessage = fmt.Sprintf("Audio saved to: %s (%d bytes).", savedFilename, len(audioBytes))
				log.Printf(fileSaveMessage)
			}
		}
	} else {
		base64AudioData := base64.StdEncoding.EncodeToString(audioBytes)
		contentItems = append(contentItems, mcp.AudioContent{Type: "audio", Data: base64AudioData, MIMEType: mimeType})
		fileSaveMessage = "Audio data is included in the response."
	}

	resultText := fmt.Sprintf("Speech synthesized successfully with voice %s. %s", voiceName, fileSaveMessage)
	contentItems = append([]mcp.Content{mcp.TextContent{Type: "text", Text: resultText}}, contentItems...)

	return &mcp.CallToolResult{Content: contentItems}, nil
}

// --- API Helper Function ---

func callGeminiTTSAPIWithSDK(ctx context.Context, text, prompt, voiceName, modelName, audioEncoding, languageCode string) ([]byte, error) {
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create texttospeech client: %w", err)
	}
	defer client.Close()

	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: languageCode,
			Name:         voiceName,
			ModelName:    modelName,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding(texttospeechpb.AudioEncoding_value[audioEncoding]),
		},
	}

	if prompt != "" {
		req.Input.Prompt = &prompt
	}

	resp, err := client.SynthesizeSpeech(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize speech: %w", err)
	}

	return resp.AudioContent, nil
}
