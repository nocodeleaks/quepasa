package models

// Informações básicas sobre mensagens no formato QuePasa
// Utilizada na API do QuePasa para troca com outros sistemas
type QPMessageInfo struct {
	ID        string `json:"id"`
	Timestamp uint64 `json:"timestamp"`
}