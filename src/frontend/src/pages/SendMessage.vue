<template>
  <div class="send-page">
    <div class="page-header">
      <div class="header-content">
        <button @click="$router.back()" class="back-link">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
          </svg>
          Voltar
        </button>
        <h1>Enviar Mensagem</h1>
        <p>Envie mensagens via WhatsApp</p>
      </div>
    </div>

    <div class="content-grid">
      <!-- Form -->
      <div class="send-card">
        <div v-if="error" class="error-box">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
          </svg>
          <span>{{ error }}</span>
        </div>

        <div v-if="success" class="success-box">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L18 9l-9 9z"/>
          </svg>
          <span>{{ success }}</span>
        </div>

        <form @submit.prevent="sendMessage" class="send-form">
          <!-- Recipient -->
          <div class="form-group">
            <label for="recipient">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
              </svg>
              Destinatário
            </label>
            <div class="input-wrapper">
              <span class="input-prefix">+</span>
              <input 
                id="recipient"
                v-model="recipient"
                type="tel"
                class="form-input with-prefix"
                placeholder="5511999999999"
                required
              />
              <button type="button" class="btn-small contact-search-btn" @click="openContactSearch()">Buscar</button>
            </div>
            <div v-if="contactSearchVisible" class="contact-search">
              <div class="contact-search-controls">
                <input v-model="contactSearchQuery" @keyup.enter="searchContacts()" placeholder="Pesquisar nome ou telefone" class="form-input" />
                <button type="button" class="btn-small" @click="searchContacts()">Buscar</button>
                <button type="button" class="btn-small" @click="closeContactSearch()">Fechar</button>
              </div>
              <div class="contact-search-list">
                <div v-if="contactSearchLoading">Buscando...</div>
                <div v-else-if="!contactSearchLoading && contactSearchResults.length === 0">Nenhum contato encontrado</div>
                <ul v-else>
                  <li v-for="ct in contactSearchResults" :key="ct.id">
                    <button type="button" class="contact-result" @click="selectContact(ct)">{{ ct.title || ct.id || ct.phone }} — {{ ct.phone }}</button>
                  </li>
                </ul>
              </div>
            </div>
            <span class="input-hint">Número completo com código do país (sem espaços ou traços)</span>
          </div>

          <!-- Message Type Tabs -->
          <div class="form-group">
            <label>Tipo de Mensagem</label>
            <div class="message-tabs">
              <button type="button" @click="msgType = 'text'" :class="{ active: msgType === 'text' }">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/>
                </svg>
                Texto
              </button>
              <button type="button" @click="msgType = 'image'" :class="{ active: msgType === 'image' }">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z"/>
                </svg>
                Imagem
              </button>
              <button type="button" @click="msgType = 'document'" :class="{ active: msgType === 'document' }">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>
                </svg>
                Documento
              </button>
              <button type="button" @click="msgType = 'audio'" :class="{ active: msgType === 'audio' }">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M12 14c1.66 0 2.99-1.34 2.99-3L15 5c0-1.66-1.34-3-3-3S9 3.34 9 5v6c0 1.66 1.34 3 3 3zm5.3-3c0 3-2.54 5.1-5.3 5.1S6.7 14 6.7 11H5c0 3.41 2.72 6.23 6 6.72V21h2v-3.28c3.28-.48 6-3.3 6-6.72h-1.7z"/>
                </svg>
                Áudio
              </button>
            </div>
          </div>

          <!-- Text Message -->
          <div v-if="msgType === 'text'" class="form-group">
            <label for="text">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M20 2H4c-1.1 0-1.99.9-1.99 2L2 22l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm-2 12H6v-2h12v2zm0-3H6V9h12v2zm0-3H6V6h12v2z"/>
              </svg>
              Mensagem
            </label>
            <textarea 
              id="text"
              v-model="text"
              class="form-textarea"
              placeholder="Digite sua mensagem..."
              rows="5"
              required
            ></textarea>
            <span class="input-hint">{{ text.length }} caracteres</span>
          </div>

          <!-- Image Options -->
          <div v-if="msgType === 'image'" class="form-group">
            <label>Origem da Imagem</label>
            <div class="source-tabs">
              <button type="button" @click="mediaSource = 'file'" :class="{ active: mediaSource === 'file' }">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M9 16h6v-6h4l-7-7-7 7h4v6zm-4 2h14v2H5v-2z"/>
                </svg>
                Upload de Arquivo
              </button>
              <button type="button" @click="mediaSource = 'url'" :class="{ active: mediaSource === 'url' }">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M3.9 12c0-1.71 1.39-3.1 3.1-3.1h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-1.9H7c-1.71 0-3.1-1.39-3.1-3.1zM8 13h8v-2H8v2zm9-6h-4v1.9h4c1.71 0 3.1 1.39 3.1 3.1s-1.39 3.1-3.1 3.1h-4V17h4c2.76 0 5-2.24 5-5s-2.24-5-5-5z"/>
                </svg>
                URL
              </button>
            </div>
          </div>

          <div v-if="msgType === 'image' && mediaSource === 'file'" class="form-group">
            <label>
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z"/>
              </svg>
              Selecionar Imagem
            </label>
            <div class="file-upload" @click="imageInput?.click()" :class="{ 'has-file': selectedFile }">
              <input ref="imageInput" type="file" accept="image/*" @change="handleFileSelect" hidden />
              <div v-if="!selectedFile" class="upload-placeholder">
                <svg viewBox="0 0 24 24" width="40" height="40" fill="currentColor">
                  <path d="M19.35 10.04C18.67 6.59 15.64 4 12 4 9.11 4 6.6 5.64 5.35 8.04 2.34 8.36 0 10.91 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96zM14 13v4h-4v-4H7l5-5 5 5h-3z"/>
                </svg>
                <span>Clique para selecionar uma imagem</span>
                <span class="file-hint">JPG, PNG, GIF - Máx 16MB</span>
              </div>
              <div v-else class="file-selected">
                <svg viewBox="0 0 24 24" width="32" height="32" fill="currentColor">
                  <path d="M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z"/>
                </svg>
                <div class="file-info">
                  <span class="file-name">{{ selectedFile.name }}</span>
                  <span class="file-size">{{ formatFileSize(selectedFile.size) }}</span>
                </div>
                <button type="button" @click.stop="clearFile" class="file-remove">✕</button>
              </div>
            </div>
          </div>

          <div v-if="msgType === 'image' && mediaSource === 'url'" class="form-group">
            <label for="attachmentUrl">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M3.9 12c0-1.71 1.39-3.1 3.1-3.1h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-1.9H7c-1.71 0-3.1-1.39-3.1-3.1zM8 13h8v-2H8v2zm9-6h-4v1.9h4c1.71 0 3.1 1.39 3.1 3.1s-1.39 3.1-3.1 3.1h-4V17h4c2.76 0 5-2.24 5-5s-2.24-5-5-5z"/>
              </svg>
              URL da Imagem
            </label>
            <input 
              id="attachmentUrl"
              v-model="attachmentUrl"
              type="url"
              class="form-input"
              placeholder="https://exemplo.com/imagem.jpg"
            />
          </div>

          <!-- Caption for image -->
          <div v-if="msgType === 'image'" class="form-group">
            <label for="caption">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25z"/>
              </svg>
              Legenda (opcional)
            </label>
            <input 
              id="caption"
              v-model="text"
              type="text"
              class="form-input"
              placeholder="Legenda da imagem..."
            />
          </div>

          <!-- Document Options -->
          <div v-if="msgType === 'document'" class="form-group">
            <label>Origem do Documento</label>
            <div class="source-tabs">
              <button type="button" @click="mediaSource = 'file'" :class="{ active: mediaSource === 'file' }">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M9 16h6v-6h4l-7-7-7 7h4v6zm-4 2h14v2H5v-2z"/>
                </svg>
                Upload de Arquivo
              </button>
              <button type="button" @click="mediaSource = 'url'" :class="{ active: mediaSource === 'url' }">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M3.9 12c0-1.71 1.39-3.1 3.1-3.1h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-1.9H7c-1.71 0-3.1-1.39-3.1-3.1zM8 13h8v-2H8v2zm9-6h-4v1.9h4c1.71 0 3.1 1.39 3.1 3.1s-1.39 3.1-3.1 3.1h-4V17h4c2.76 0 5-2.24 5-5s-2.24-5-5-5z"/>
                </svg>
                URL
              </button>
            </div>
          </div>

          <div v-if="msgType === 'document' && mediaSource === 'file'" class="form-group">
            <label>
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6z"/>
              </svg>
              Selecionar Documento
            </label>
            <div class="file-upload" @click="docInput?.click()" :class="{ 'has-file': selectedFile }">
              <input ref="docInput" type="file" @change="handleFileSelect" hidden />
              <div v-if="!selectedFile" class="upload-placeholder">
                <svg viewBox="0 0 24 24" width="40" height="40" fill="currentColor">
                  <path d="M19.35 10.04C18.67 6.59 15.64 4 12 4 9.11 4 6.6 5.64 5.35 8.04 2.34 8.36 0 10.91 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96zM14 13v4h-4v-4H7l5-5 5 5h-3z"/>
                </svg>
                <span>Clique para selecionar um documento</span>
                <span class="file-hint">PDF, DOC, XLS, etc - Máx 100MB</span>
              </div>
              <div v-else class="file-selected">
                <svg viewBox="0 0 24 24" width="32" height="32" fill="currentColor">
                  <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>
                </svg>
                <div class="file-info">
                  <span class="file-name">{{ selectedFile.name }}</span>
                  <span class="file-size">{{ formatFileSize(selectedFile.size) }}</span>
                </div>
                <button type="button" @click.stop="clearFile" class="file-remove">✕</button>
              </div>
            </div>
          </div>

          <div v-if="msgType === 'document' && mediaSource === 'url'" class="form-group">
            <label for="attachmentUrl">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M3.9 12c0-1.71 1.39-3.1 3.1-3.1h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-1.9H7c-1.71 0-3.1-1.39-3.1-3.1zM8 13h8v-2H8v2zm9-6h-4v1.9h4c1.71 0 3.1 1.39 3.1 3.1s-1.39 3.1-3.1 3.1h-4V17h4c2.76 0 5-2.24 5-5s-2.24-5-5-5z"/>
              </svg>
              URL do Documento
            </label>
            <input 
              id="attachmentUrl"
              v-model="attachmentUrl"
              type="url"
              class="form-input"
              placeholder="https://exemplo.com/documento.pdf"
            />
          </div>

          <!-- Filename for document -->
          <div v-if="msgType === 'document'" class="form-group">
            <label for="filename">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M14 2H6c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V8l-6-6z"/>
              </svg>
              Nome do Arquivo (opcional)
            </label>
            <input 
              id="filename"
              v-model="filename"
              type="text"
              class="form-input"
              placeholder="documento.pdf"
            />
          </div>

          <!-- Audio Options -->
          <div v-if="msgType === 'audio'" class="form-group">
            <label>Origem do Áudio</label>
            <div class="source-tabs">
              <button type="button" @click="mediaSource = 'record'" :class="{ active: mediaSource === 'record' }">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M12 14c1.66 0 2.99-1.34 2.99-3L15 5c0-1.66-1.34-3-3-3S9 3.34 9 5v6c0 1.66 1.34 3 3 3zm5.3-3c0 3-2.54 5.1-5.3 5.1S6.7 14 6.7 11H5c0 3.41 2.72 6.23 6 6.72V21h2v-3.28c3.28-.48 6-3.3 6-6.72h-1.7z"/>
                </svg>
                Gravar Áudio
              </button>
              <button type="button" @click="mediaSource = 'file'" :class="{ active: mediaSource === 'file' }">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M9 16h6v-6h4l-7-7-7 7h4v6zm-4 2h14v2H5v-2z"/>
                </svg>
                Upload de Arquivo
              </button>
              <button type="button" @click="mediaSource = 'url'" :class="{ active: mediaSource === 'url' }">
                <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                  <path d="M3.9 12c0-1.71 1.39-3.1 3.1-3.1h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-1.9H7c-1.71 0-3.1-1.39-3.1-3.1zM8 13h8v-2H8v2zm9-6h-4v1.9h4c1.71 0 3.1 1.39 3.1 3.1s-1.39 3.1-3.1 3.1h-4V17h4c2.76 0 5-2.24 5-5s-2.24-5-5-5z"/>
                </svg>
                URL
              </button>
            </div>
          </div>

          <!-- Audio Recording -->
          <div v-if="msgType === 'audio' && mediaSource === 'record'" class="form-group">
            <div class="audio-recorder" :class="{ recording: isRecording }">
              <div v-if="!isRecording && !audioBlob" class="recorder-idle">
                <button type="button" @click="startRecording" class="btn-record">
                  <svg viewBox="0 0 24 24" width="32" height="32" fill="currentColor">
                    <path d="M12 14c1.66 0 2.99-1.34 2.99-3L15 5c0-1.66-1.34-3-3-3S9 3.34 9 5v6c0 1.66 1.34 3 3 3z"/>
                  </svg>
                </button>
                <span>Clique para gravar</span>
              </div>
              <div v-else-if="isRecording" class="recorder-active">
                <div class="recording-indicator">
                  <span class="recording-dot"></span>
                  <span class="recording-time">{{ recordingTime }}</span>
                </div>
                <button type="button" @click="stopRecording" class="btn-stop">
                  <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
                    <path d="M6 6h12v12H6z"/>
                  </svg>
                  Parar
                </button>
              </div>
              <div v-else class="recorder-done">
                <audio :src="audioUrl" controls class="audio-preview"></audio>
                <button type="button" @click="clearAudio" class="btn-clear">
                  <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                    <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
                  </svg>
                  Limpar
                </button>
              </div>
            </div>
          </div>

          <div v-if="msgType === 'audio' && mediaSource === 'file'" class="form-group">
            <label>
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z"/>
              </svg>
              Selecionar Áudio
            </label>
            <div class="file-upload" @click="audioInput?.click()" :class="{ 'has-file': selectedFile }">
              <input ref="audioInput" type="file" accept="audio/*" @change="handleFileSelect" hidden />
              <div v-if="!selectedFile" class="upload-placeholder">
                <svg viewBox="0 0 24 24" width="40" height="40" fill="currentColor">
                  <path d="M19.35 10.04C18.67 6.59 15.64 4 12 4 9.11 4 6.6 5.64 5.35 8.04 2.34 8.36 0 10.91 0 14c0 3.31 2.69 6 6 6h13c2.76 0 5-2.24 5-5 0-2.64-2.05-4.78-4.65-4.96zM14 13v4h-4v-4H7l5-5 5 5h-3z"/>
                </svg>
                <span>Clique para selecionar um áudio</span>
                <span class="file-hint">MP3, OGG, WAV - Máx 16MB</span>
              </div>
              <div v-else class="file-selected">
                <svg viewBox="0 0 24 24" width="32" height="32" fill="currentColor">
                  <path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z"/>
                </svg>
                <div class="file-info">
                  <span class="file-name">{{ selectedFile.name }}</span>
                  <span class="file-size">{{ formatFileSize(selectedFile.size) }}</span>
                </div>
                <button type="button" @click.stop="clearFile" class="file-remove">✕</button>
              </div>
            </div>
          </div>

          <div v-if="msgType === 'audio' && mediaSource === 'url'" class="form-group">
            <label for="attachmentUrl">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M3.9 12c0-1.71 1.39-3.1 3.1-3.1h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-1.9H7c-1.71 0-3.1-1.39-3.1-3.1zM8 13h8v-2H8v2zm9-6h-4v1.9h4c1.71 0 3.1 1.39 3.1 3.1s-1.39 3.1-3.1 3.1h-4V17h4c2.76 0 5-2.24 5-5s-2.24-5-5-5z"/>
              </svg>
              URL do Áudio
            </label>
            <input 
              id="attachmentUrl"
              v-model="attachmentUrl"
              type="url"
              class="form-input"
              placeholder="https://exemplo.com/audio.mp3"
            />
          </div>

          <button type="submit" class="btn-send" :disabled="sending">
            <span v-if="sending" class="spinner"></span>
            <svg v-else viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
              <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
            </svg>
            {{ sending ? 'Enviando...' : 'Enviar Mensagem' }}
          </button>
        </form>
      </div>

      <!-- Preview / History -->
      <div class="preview-card">
        <h3>
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z"/>
          </svg>
          Preview
        </h3>

        <div class="phone-mockup">
          <div class="phone-header">
            <div class="contact-info">
              <div class="contact-avatar">{{ getInitial(recipient) }}</div>
              <span>+{{ recipient || '...' }}</span>
            </div>
          </div>
          <div class="phone-messages">
            <div class="message-bubble sent">
              <div v-if="msgType === 'text'" class="bubble-text">{{ text || 'Sua mensagem aparecerá aqui...' }}</div>
              <div v-else-if="msgType === 'image'" class="bubble-media">
                <div class="media-placeholder">
                  <svg viewBox="0 0 24 24" width="32" height="32" fill="currentColor">
                    <path d="M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z"/>
                  </svg>
                </div>
                <div v-if="text" class="bubble-text">{{ text }}</div>
              </div>
              <div v-else class="bubble-media">
                <div class="media-placeholder file">
                  <svg viewBox="0 0 24 24" width="32" height="32" fill="currentColor">
                    <path d="M14 2H6c-1.1 0-2 .9-2 2v16c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V8l-6-6z"/>
                  </svg>
                  <span>{{ filename || 'arquivo' }}</span>
                </div>
              </div>
              <div class="bubble-time">{{ currentTime }}</div>
            </div>
          </div>
        </div>

        <!-- Recent sends -->
        <div v-if="recentSends.length > 0" class="recent-sends">
          <h4>Envios Recentes</h4>
          <div class="recent-list">
            <div v-for="(send, i) in recentSends" :key="i" class="recent-item" :class="{ failed: send.error }">
              <span class="recent-to">+{{ send.to }}</span>
              <span class="recent-status">{{ send.error ? 'Falha' : 'Enviado' }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, computed, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/services/api'
import { pushToast } from '@/services/toast'

export default defineComponent({
  setup() {
    const route = useRoute()
    const token = route.params.token as string

    const recipient = ref('')
    const text = ref('')
    const msgType = ref<'text' | 'image' | 'document' | 'audio'>('text')
    const mediaSource = ref<'file' | 'url' | 'record'>('file')
    const attachmentUrl = ref('')
    const filename = ref('')
    const sending = ref(false)
    const error = ref('')
    const success = ref('')
    const recentSends = ref<{ to: string; error?: boolean }[]>([])
    
    // Contact search state
    const contactSearchQuery = ref('')
    const contactSearchResults = ref<any[]>([])
    const contactSearchVisible = ref(false)
    const contactSearchLoading = ref(false)
    const contactSearchError = ref('')

    function openContactSearch() {
      contactSearchVisible.value = true
      contactSearchQuery.value = ''
      contactSearchResults.value = []
    }

    function closeContactSearch() {
      contactSearchVisible.value = false
    }

    async function searchContacts() {
      if (!contactSearchQuery.value) {
        pushToast('Digite nome ou telefone para buscar', 'error')
        return
      }
      contactSearchLoading.value = true
      contactSearchError.value = ''
      try {
        const res = await api.post('/api/contact/search', { query: contactSearchQuery.value })
        contactSearchResults.value = res.data?.contacts || []
        if (!contactSearchResults.value.length) pushToast('Nenhum contato encontrado', 'info')
      } catch (e: any) {
        contactSearchError.value = e?.response?.data?.result || e?.message || 'Erro ao buscar contatos'
        pushToast(contactSearchError.value, 'error')
      } finally {
        contactSearchLoading.value = false
      }
    }

    function selectContact(ct: any) {
      if (!ct) return
      recipient.value = ct.phone || ct.id || ct.lid || ''
      contactSearchVisible.value = false
      pushToast(`Destinatário definido: ${recipient.value}`, 'success')
    }

    // File upload
    const selectedFile = ref<File | null>(null)
    const imageInput = ref<HTMLInputElement | null>(null)
    const docInput = ref<HTMLInputElement | null>(null)
    const audioInput = ref<HTMLInputElement | null>(null)
    
    // Audio recording
    const isRecording = ref(false)
    const recordingTime = ref('0:00')
    const audioBlob = ref<Blob | null>(null)
    const audioUrl = ref('')
    let mediaRecorder: MediaRecorder | null = null
    let audioChunks: Blob[] = []
    let recordingInterval: number | null = null
    let recordingStartTime = 0

    const currentTime = computed(() => {
      const now = new Date()
      return now.toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })
    })

    // Reset mediaSource when changing msgType
    watch(msgType, (newType) => {
      if (newType === 'audio') {
        mediaSource.value = 'record'
      } else if (newType !== 'text') {
        mediaSource.value = 'file'
      }
      clearFile()
      clearAudio()
      attachmentUrl.value = ''
    })

    function getInitial(phone: string) {
      return phone ? phone.charAt(0) : '?'
    }
    
    function formatFileSize(bytes: number): string {
      if (bytes === 0) return '0 Bytes'
      const k = 1024
      const sizes = ['Bytes', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    }
    
    function handleFileSelect(event: Event) {
      const input = event.target as HTMLInputElement
      if (input.files && input.files[0]) {
        selectedFile.value = input.files[0]
        if (msgType.value === 'document' && !filename.value) {
          filename.value = selectedFile.value.name
        }
      }
    }
    
    function clearFile() {
      selectedFile.value = null
    }
    
    async function startRecording() {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
        mediaRecorder = new MediaRecorder(stream)
        audioChunks = []
        
        mediaRecorder.ondataavailable = (e) => {
          audioChunks.push(e.data)
        }
        
        mediaRecorder.onstop = () => {
          const blob = new Blob(audioChunks, { type: 'audio/ogg; codecs=opus' })
          audioBlob.value = blob
          audioUrl.value = URL.createObjectURL(blob)
          stream.getTracks().forEach(t => t.stop())
        }
        
        mediaRecorder.start()
        isRecording.value = true
        recordingStartTime = Date.now()
        
        recordingInterval = window.setInterval(() => {
          const elapsed = Math.floor((Date.now() - recordingStartTime) / 1000)
          const min = Math.floor(elapsed / 60)
          const sec = elapsed % 60
          recordingTime.value = `${min}:${sec.toString().padStart(2, '0')}`
        }, 1000)
      } catch (err) {
        error.value = 'Não foi possível acessar o microfone'
      }
    }
    
    function stopRecording() {
      if (mediaRecorder && isRecording.value) {
        mediaRecorder.stop()
        isRecording.value = false
        if (recordingInterval) {
          clearInterval(recordingInterval)
          recordingInterval = null
        }
      }
    }
    
    function clearAudio() {
      audioBlob.value = null
      if (audioUrl.value) {
        URL.revokeObjectURL(audioUrl.value)
        audioUrl.value = ''
      }
      recordingTime.value = '0:00'
    }
    
    onUnmounted(() => {
      if (recordingInterval) clearInterval(recordingInterval)
      if (audioUrl.value) URL.revokeObjectURL(audioUrl.value)
    })

    async function sendMessage() {
      sending.value = true
      error.value = ''
      success.value = ''

      try {
        if (msgType.value === 'text') {
          // Use SPA endpoint for text messages
          const payload = {
            recipient: recipient.value,
            message: text.value
          }
          await api.post(`/api/server/${token}/send`, payload)
        } else {
          // For media types, use the bot API
          const hasFile = selectedFile.value || audioBlob.value
          
          if (mediaSource.value !== 'url' && hasFile) {
            // Upload file using FormData
            const formData = new FormData()
            formData.append('chatid', recipient.value)
            
            if (audioBlob.value) {
              formData.append('attach', audioBlob.value, 'audio.ogg')
            } else if (selectedFile.value) {
              formData.append('attach', selectedFile.value)
            }
            
            if (text.value) formData.append('text', text.value)
            if (filename.value) formData.append('filename', filename.value)

            await api.post(`/bot/${token}/sendattach`, formData, {
              headers: { 
                'X-QUEPASA-TOKEN': token,
                'Content-Type': 'multipart/form-data'
              }
            })
          } else {
            // Use URL
            const mediaPayload: any = {
              chatid: recipient.value,
              url: attachmentUrl.value
            }
            if (text.value) mediaPayload.text = text.value
            if (filename.value) mediaPayload.filename = filename.value

            await api.post(`/bot/${token}/send`, mediaPayload, {
              headers: { 'X-QUEPASA-TOKEN': token }
            })
          }
        }

        success.value = `Mensagem enviada para +${recipient.value}`
        recentSends.value.unshift({ to: recipient.value })
        if (recentSends.value.length > 5) recentSends.value.pop()

        // Clear form
        text.value = ''
        attachmentUrl.value = ''
        filename.value = ''
        clearFile()
        clearAudio()
      } catch (err: any) {
        error.value = err?.response?.data?.result || err.message || 'Erro ao enviar mensagem'
        recentSends.value.unshift({ to: recipient.value, error: true })
        if (recentSends.value.length > 5) recentSends.value.pop()
      } finally {
        sending.value = false
      }
    }

    return {
      token, recipient, text, msgType, mediaSource, attachmentUrl, filename,
      sending, error, success, recentSends, currentTime,
      selectedFile, isRecording, recordingTime, audioBlob, audioUrl,
      imageInput, docInput, audioInput,
      getInitial, sendMessage, formatFileSize, handleFileSelect, clearFile,
      startRecording, stopRecording, clearAudio,
      // contact search
      contactSearchQuery, contactSearchResults, contactSearchVisible, contactSearchLoading, openContactSearch, closeContactSearch, searchContacts, selectContact
    }
  }
})
</script>

<style scoped>
.send-page {
  max-width: 1100px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
}

.header-content {
  text-align: center;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #6b7280;
  text-decoration: none;
  font-size: 14px;
  margin-bottom: 16px;
}

.page-header h1 {
  font-size: 28px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 8px;
}

.page-header p {
  color: #6b7280;
  margin: 0;
}

.content-grid {
  display: grid;
  grid-template-columns: 1fr 360px;
  gap: 24px;
}

@media (max-width: 900px) {
  .content-grid {
    grid-template-columns: 1fr;
  }
}

.send-card, .preview-card {
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

.error-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 12px;
  color: #dc2626;
  margin-bottom: 16px;
}

.success-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  background: #f5efff;
  border: 1px solid rgba(124, 58, 237, 0.12);
  border-radius: 12px;
  color: var(--branding-secondary, #5B21B6);
  margin-bottom: 16px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 8px;
}

.form-group label svg {
  color: #6b7280;
}

.input-wrapper {
  position: relative;
}
.contact-search {
  margin-top: 8px;
  border: 1px solid #e5e7eb;
  padding: 8px;
  border-radius: 8px;
  background: #fff;
}
.contact-search-controls {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}
.contact-search-list ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 8px;
}
.contact-result {
  width: 100%;
  text-align: left;
  padding: 8px;
  border: 1px solid #eee;
  border-radius: 6px;
  background: #f9fafb;
  cursor: pointer;
}
.contact-result:hover {
  background: #f3f4f6;
}
.contact-search-btn{
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
}

.input-prefix {
  position: absolute;
  left: 16px;
  top: 50%;
  transform: translateY(-50%);
  color: #6b7280;
  font-weight: 600;
}

.form-input {
  width: 100%;
  padding: 14px 16px;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  font-size: 16px;
  transition: all 0.2s;
}

.form-input.with-prefix {
  padding-left: 32px;
}

.form-input:focus {
  outline: none;
  border-color: var(--branding-primary, #7C3AED);
}

.form-textarea {
  width: 100%;
  padding: 14px 16px;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  font-size: 16px;
  resize: vertical;
  min-height: 120px;
}

.form-textarea:focus {
  outline: none;
  border-color: var(--branding-primary, #7C3AED);
}

.input-hint {
  display: block;
  font-size: 12px;
  color: #9ca3af;
  margin-top: 6px;
}

.message-tabs {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.message-tabs button {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  background: #f3f4f6;
  border: 2px solid transparent;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
}

.message-tabs button:hover {
  background: #e5e7eb;
}

.message-tabs button.active {
  background: #f5efff;
  border-color: var(--branding-primary, #7C3AED);
  color: var(--branding-secondary, #5B21B6);
}

.btn-send {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 16px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-send:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(124, 58, 237, 0.25);
}

.btn-send:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.preview-card h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  color: #374151;
  margin: 0 0 16px;
}

.phone-mockup {
  background: #e5ddd5;
  border-radius: 16px;
  overflow: hidden;
}

.phone-header {
  background: #075e54;
  padding: 12px 16px;
}

.contact-info {
  display: flex;
  align-items: center;
  gap: 10px;
  color: white;
}

.contact-avatar {
  width: 36px;
  height: 36px;
  background: #128c7e;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
}

.phone-messages {
  padding: 16px;
  min-height: 200px;
}

.message-bubble {
  max-width: 80%;
  margin-left: auto;
  background: #dcf8c6;
  padding: 8px 12px;
  border-radius: 8px 0 8px 8px;
  position: relative;
}

.bubble-text {
  color: #111827;
  font-size: 14px;
  line-height: 1.4;
  word-break: break-word;
}

.bubble-media {
  margin-bottom: 4px;
}

.media-placeholder {
  width: 100%;
  height: 120px;
  background: rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #6b7280;
  margin-bottom: 8px;
}

.media-placeholder.file {
  flex-direction: column;
  gap: 8px;
  height: 80px;
}

.media-placeholder.file span {
  font-size: 12px;
}

.bubble-time {
  font-size: 11px;
  color: #667781;
  text-align: right;
  margin-top: 4px;
}

.recent-sends {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #e5e7eb;
}

.recent-sends h4 {
  font-size: 14px;
  color: #6b7280;
  margin: 0 0 12px;
}

.recent-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.recent-item {
  display: flex;
  justify-content: space-between;
  padding: 8px 12px;
  background: #f0fdf4;
  border-radius: 8px;
  font-size: 13px;
}

.recent-item.failed {
  background: #fef2f2;
}

.recent-to {
  font-weight: 500;
  color: #374151;
}

.recent-status {
  color: var(--branding-secondary, #5B21B6);
}

.recent-item.failed .recent-status {
  color: #dc2626;
}

/* Source Tabs */
.source-tabs {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.source-tabs button {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  background: #f3f4f6;
  border: 2px solid transparent;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
}

.source-tabs button:hover {
  background: #e5e7eb;
}

.source-tabs button.active {
  background: #f5efff;
  border-color: var(--branding-primary, #7C3AED);
  color: var(--branding-secondary, #5B21B6);
}

/* File Upload */
.file-upload {
  border: 2px dashed #d1d5db;
  border-radius: 12px;
  padding: 24px;
  cursor: pointer;
  transition: all 0.2s;
  background: #fafafa;
}

.file-upload:hover {
  border-color: var(--branding-primary, #7C3AED);
  background: #f5efff;
}

.file-upload.has-file {
  border-style: solid;
  border-color: var(--branding-primary, #7C3AED);
  background: #f5efff;
}

.upload-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #6b7280;
}

.upload-placeholder svg {
  color: #9ca3af;
}

.file-hint {
  font-size: 12px;
  color: #9ca3af;
}

.file-selected {
  display: flex;
  align-items: center;
  gap: 12px;
}

.file-selected svg {
  color: var(--branding-primary, #7C3AED);
  flex-shrink: 0;
}

.file-info {
  flex: 1;
  min-width: 0;
}

.file-name {
  display: block;
  font-weight: 500;
  color: #374151;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-size {
  font-size: 12px;
  color: #9ca3af;
}

.file-remove {
  width: 28px;
  height: 28px;
  background: #fee2e2;
  color: #dc2626;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  transition: all 0.2s;
}

.file-remove:hover {
  background: #fecaca;
}

/* Audio Recorder */
.audio-recorder {
  background: #f8fafc;
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  padding: 24px;
  text-align: center;
}

.audio-recorder.recording {
  border-color: #ef4444;
  background: #fef2f2;
}

.recorder-idle {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.btn-record {
  width: 64px;
  height: 64px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.btn-record:hover {
  transform: scale(1.1);
  box-shadow: 0 8px 16px rgba(124, 58, 237, 0.25);
}

.recorder-active {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
}

.recording-indicator {
  display: flex;
  align-items: center;
  gap: 12px;
}

.recording-dot {
  width: 12px;
  height: 12px;
  background: #ef4444;
  border-radius: 50%;
  animation: pulse 1s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.recording-time {
  font-size: 24px;
  font-weight: 600;
  color: #ef4444;
  font-variant-numeric: tabular-nums;
}

.btn-stop {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-stop:hover {
  background: #dc2626;
}

.recorder-done {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.audio-preview {
  width: 100%;
  max-width: 280px;
}

.btn-clear {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: #fee2e2;
  color: #dc2626;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-clear:hover {
  background: #fecaca;
}
</style>
