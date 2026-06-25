package library

/*
<summary>

	Multipurpose Internet Mail Extensions
	Override default system types

</summary>
*/
var MIMEs = map[string]string{
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       ".xlsx",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
	"application/zip": ".zip",
	"audio/mpeg":      ".mp3",
	"audio/opus":      ".opus",
	"audio/wave":      ".wav",
	"audio/wav":       ".wav",
	"audio/x-wav":     ".wav",
	"image/jpeg":      ".jpeg",
	"image/webp":      ".webp",
	"text/csv":        ".csv",
	"text/xml":        ".xml",
	"text/plain":      ".txt",
	"video/mp4":       ".mp4",
}
