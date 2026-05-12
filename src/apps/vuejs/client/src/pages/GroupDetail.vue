<template>
  <div class="group-detail-page">

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>{{ t('group_detail_loading') }}</p>
    </div>

    <div v-else-if="error" class="error-full">
      <svg viewBox="0 0 24 24" width="48" height="48" fill="currentColor">
        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
      </svg>
      <p>{{ error }}</p>
      <router-link :to="`/server/${token}/groups`" class="btn-back">{{ t('group_detail_back_to_groups') }}</router-link>
    </div>

    <template v-else>
      <div class="page-header">
        <router-link :to="`/server/${token}/groups`" class="back-link">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M20 11H7.83l5.59-5.59L12 4l-8 8 8 8 1.41-1.41L7.83 13H20v-2z"/>
          </svg>
          {{ t('group_detail_back_to_groups') }}
        </router-link>
      </div>

      <div class="group-layout">
        <!-- Left: details -->
        <div class="details-col">

          <div class="profile-card">
            <div class="profile-photo-wrap">
              <img v-if="groupPicture" :src="groupPicture" :alt="group.Name" class="profile-photo" />
              <div v-else class="profile-photo-placeholder" :style="avatarStyle">
                {{ initials(group.Name) }}
              </div>
              <button v-if="isAdmin" class="edit-photo-btn" type="button" @click="openModal('photo')" :title="t('group_detail_change_photo')">
                <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
                  <path d="M12 15.2A3.2 3.2 0 0 1 8.8 12 3.2 3.2 0 0 1 12 8.8 3.2 3.2 0 0 1 15.2 12 3.2 3.2 0 0 1 12 15.2M18.2 12c0-.22-.02-.44-.05-.65l1.4-1.1c.13-.1.16-.28.08-.42l-1.33-2.3a.33.33 0 0 0-.4-.14l-1.65.66a4.87 4.87 0 0 0-1.12-.65l-.25-1.76A.33.33 0 0 0 14.6 6h-2.66a.33.33 0 0 0-.33.28l-.25 1.76a4.87 4.87 0 0 0-1.12.65l-1.65-.66a.33.33 0 0 0-.4.14L6.86 10.43c-.08.14-.05.32.08.42l1.4 1.1c-.03.21-.05.43-.05.65 0 .22.02.44.05.65l-1.4 1.1c-.13.1-.16.28-.08.42l1.33 2.3c.08.14.27.18.4.14l1.65-.66c.35.25.72.47 1.12.65l.25 1.76c.04.17.18.28.33.28h2.66c.15 0 .3-.11.33-.28l.25-1.76a4.87 4.87 0 0 0 1.12-.65l1.65.66c.13.04.32 0 .4-.14l1.33-2.3c.08-.14.05-.32-.08-.42l-1.4-1.1c.03-.21.05-.43.05-.65z"/>
                </svg>
              </button>
            </div>

            <div class="profile-info">
              <div class="profile-name-row">
                <h1 class="profile-name">{{ group.Name || t('groups_unnamed') }}</h1>
                <button v-if="isAdmin" class="edit-inline-btn" type="button" @click="openModal('name')" :title="t('group_detail_change_name')">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
                    <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/>
                  </svg>
                </button>
              </div>
              <p class="profile-meta">{{ t('group_detail_meta').replace('{0}', String(group.Participants?.length || 0)) }}</p>
            </div>
          </div>

          <!-- Description -->
          <div v-if="group.Topic || isAdmin" class="info-card">
            <div class="card-header">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z"/>
              </svg>
              <span>{{ t('group_detail_description') }}</span>
              <button v-if="isAdmin" class="edit-inline-btn" type="button" @click="openModal('topic')">
                <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor">
                  <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/>
                </svg>
              </button>
            </div>
            <p v-if="group.Topic" class="description-text">{{ group.Topic }}</p>
            <p v-else class="description-empty">{{ t('group_detail_no_description') }}</p>
            <p v-if="group.TopicSetAt" class="description-meta">{{ t('group_detail_created_at').replace('{0}', formatDate(group.TopicSetAt)) }}</p>
          </div>

          <!-- Actions row -->
          <div class="actions-row">
            <button class="action-btn" type="button" @click="openModal('invite')">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M3.9 12c0-1.71 1.39-3.1 3.1-3.1h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-1.9H7c-1.71 0-3.1-1.39-3.1-3.1zM8 13h8v-2H8v2zm9-6h-4v1.9h4c1.71 0 3.1 1.39 3.1 3.1s-1.39 3.1-3.1 3.1h-4V17h4c2.76 0 5-2.24 5-5s-2.24-5-5-5z"/>
              </svg>
              {{ t('group_detail_invite_link') }}
            </button>
            <button v-if="isAdmin" class="action-btn" type="button" @click="openModal('addParticipant')">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M15 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm-9-2V7H4v3H1v2h3v3h2v-3h3v-2H6zm9 4c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
              </svg>
              {{ t('group_detail_add') }}
            </button>
            <button v-if="isAdmin" class="action-btn" type="button" @click="openModal('requests')">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
              </svg>
              {{ t('group_detail_join_requests') }}
              <span v-if="pendingRequests.length" class="req-badge">{{ pendingRequests.length }}</span>
            </button>
          </div>

          <!-- Members -->
          <div class="info-card members-card">
            <div class="card-header">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
              </svg>
              <span>{{ t('group_detail_members_count').replace('{0}', String(group.Participants?.length || 0)) }}</span>
              <button class="edit-inline-btn" type="button" @click="showSearch = !showSearch">
                <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor">
                  <path d="M15.5 14h-.79l-.28-.27C15.41 12.59 16 11.11 16 9.5 16 5.91 13.09 3 9.5 3S3 5.91 3 9.5 5.91 16 9.5 16c1.61 0 3.09-.59 4.23-1.57l.27.28v.79l5 4.99L20.49 19l-4.99-5zm-6 0C7.01 14 5 11.99 5 9.5S7.01 5 9.5 5 14 7.01 14 9.5 11.99 14 9.5 14z"/>
                </svg>
              </button>
            </div>
            <div v-if="showSearch" class="member-search-wrap">
              <input v-model="participantSearch" type="text" class="member-search" :placeholder="t('group_detail_search_members')" />
            </div>
            <div class="participants-list">
              <div v-for="p in filteredParticipants" :key="participantUniqueKey(p)" class="participant-row">
                <div class="participant-avatar">
                  <img v-if="participantPictures[participantPrimaryId(p)]" :src="participantPictures[participantPrimaryId(p)]" :alt="p.DisplayName" />
                  <div v-else class="participant-avatar-placeholder">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                      <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
                    </svg>
                  </div>
                </div>
                <div class="participant-info">
                  <span class="participant-name">
                    <template v-if="p.DisplayName">~ {{ p.DisplayName }}</template>
                    <template v-else>{{ formatPhone(participantPrimaryId(p)) }}</template>
                  </span>
                  <span v-if="p.DisplayName && participantPrimaryId(p)" class="participant-phone">{{ formatPhone(participantPrimaryId(p)) }}</span>
                </div>
                <div class="participant-role">
                  <span v-if="p.IsSuperAdmin" class="role-badge role-owner">{{ t('group_detail_owner') }}</span>
                  <span v-else-if="p.IsAdmin" class="role-badge role-admin">{{ t('group_detail_admin') }}</span>
                </div>
                <div v-if="isAdmin && !p.IsSuperAdmin" class="participant-actions">
                  <button class="participant-action-btn" type="button" @click="openParticipantMenu(p)" :title="t('group_detail_participant_options')">
                    <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
                      <path d="M12 8c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2zm0 2c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm0 6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2z"/>
                    </svg>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Leave / Danger -->
          <div class="danger-card">
            <button class="danger-btn" type="button" @click="openModal('leave')">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor">
                <path d="M17 7l-1.41 1.41L18.17 11H8v2h10.17l-2.58 2.58L17 17l5-5zM4 5h8V3H4c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h8v-2H4V5z"/>
              </svg>
              {{ t('group_detail_leave_group') }}
            </button>
          </div>

        </div>

        <!-- Right: messages history -->
        <div class="messages-col">
          <div class="messages-card">
            <div class="messages-header">
              <div>
                <p class="eyebrow">{{ t('group_detail_recent_history') }}</p>
                <h2>{{ t('messages') }}</h2>
              </div>
              <small v-if="totalMessages">{{ t('group_detail_messages_count').replace('{0}', String(visibleMessages.length)).replace('{1}', String(totalMessages)) }}</small>
            </div>
            <div v-if="visibleMessages.length" class="messages-list" ref="messagesListRef" @scroll.passive="onMessagesScroll">
              <div v-for="msg in visibleMessages" :key="msg.id" class="message-item">
                <div class="msg-avatar">
                  <img v-if="participantPictures[msg.participant?.phone || msg.participant?.id]" :src="participantPictures[msg.participant?.phone || msg.participant?.id]" />
                  <div v-else class="msg-avatar-placeholder">
                    <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
                      <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
                    </svg>
                  </div>
                </div>
                <div class="msg-body">
                  <div class="msg-meta">
                    <span class="msg-sender">{{ msg.participant?.title || formatPhone(msg.participant?.phone || msg.participant?.id) }}</span>
                    <span class="msg-time">{{ formatTime(msg.timestamp) }}</span>
                  </div>
                  <p class="msg-text">{{ messagePreview(msg) }}</p>
                </div>
              </div>
              <div class="load-more-area">
                <div v-if="isLoadingMore" class="load-more-info">{{ t('group_detail_loading_more') }}</div>
                <button v-else-if="hasMoreMessages" class="load-more-btn" type="button" @click="loadMoreMessages">{{ t('group_detail_load_more_50') }}</button>
                <div v-else class="load-more-info muted">{{ t('group_detail_history_start') }}</div>
              </div>
            </div>
            <div v-else class="messages-empty">{{ t('groups_no_recent_messages') }}</div>
          </div>
        </div>
      </div>
    </template>

    <!-- ===== MODALS ===== -->

    <!-- Edit Name -->
    <div v-if="modal === 'name'" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card">
        <div class="modal-header">
          <div class="modal-icon">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
              <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/>
            </svg>
          </div>
          <h3>{{ t('group_detail_change_name') }}</h3>
          <button class="modal-close" type="button" @click="closeModal">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="field">
            <label class="field-label">{{ t('group_detail_new_name') }}</label>
            <input v-model="formData.name" ref="firstInputRef" type="text" class="field-input" maxlength="25" @keydown.enter="submitModal" :placeholder="t('group_detail_name_placeholder')" />
            <span class="field-hint">{{ formData.name.length }} / 25</span>
          </div>
          <div v-if="modalError" class="modal-error">{{ modalError }}</div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" type="button" @click="closeModal" :disabled="submitting">{{ t('cancel') }}</button>
          <button class="btn-confirm" type="button" @click="submitModal" :disabled="submitting || !formData.name.trim()">
            <span v-if="submitting" class="spin-icon">⟳</span>
            {{ submitting ? t('saving') : t('save') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Edit Topic/Description -->
    <div v-if="modal === 'topic'" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card">
        <div class="modal-header">
          <div class="modal-icon">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
              <path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/>
            </svg>
          </div>
          <h3>{{ t('group_detail_description') }}</h3>
          <button class="modal-close" type="button" @click="closeModal">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="field">
            <label class="field-label">{{ t('group_detail_description') }}</label>
            <textarea v-model="formData.topic" ref="firstInputRef" class="field-input field-textarea" rows="4" :placeholder="t('group_detail_topic_placeholder')"></textarea>
          </div>
          <div v-if="modalError" class="modal-error">{{ modalError }}</div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" type="button" @click="closeModal" :disabled="submitting">{{ t('cancel') }}</button>
          <button class="btn-confirm" type="button" @click="submitModal" :disabled="submitting">
            {{ submitting ? t('saving') : t('save') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Edit Photo -->
    <div v-if="modal === 'photo'" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card">
        <div class="modal-header">
          <div class="modal-icon">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
              <path d="M21 19V5c0-1.1-.9-2-2-2H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2zM8.5 13.5l2.5 3.01L14.5 12l4.5 6H5l3.5-4.5z"/>
            </svg>
          </div>
          <h3>{{ t('group_detail_change_photo') }}</h3>
          <button class="modal-close" type="button" @click="closeModal">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="field">
            <label class="field-label">{{ t('group_detail_photo_url') }}</label>
            <input v-model="formData.photoUrl" ref="firstInputRef" type="url" class="field-input" @keydown.enter="submitModal" :placeholder="t('group_detail_photo_url_placeholder')" />
          </div>
          <div v-if="modalError" class="modal-error">{{ modalError }}</div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" type="button" @click="closeModal" :disabled="submitting">{{ t('cancel') }}</button>
          <button class="btn-confirm" type="button" @click="submitModal" :disabled="submitting || !formData.photoUrl.trim()">
            {{ submitting ? t('saving') : t('save') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Invite link -->
    <div v-if="modal === 'invite'" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card">
        <div class="modal-header">
          <div class="modal-icon">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
              <path d="M3.9 12c0-1.71 1.39-3.1 3.1-3.1h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-1.9H7c-1.71 0-3.1-1.39-3.1-3.1zM8 13h8v-2H8v2zm9-6h-4v1.9h4c1.71 0 3.1 1.39 3.1 3.1s-1.39 3.1-3.1 3.1h-4V17h4c2.76 0 5-2.24 5-5s-2.24-5-5-5z"/>
            </svg>
          </div>
          <h3>{{ t('group_detail_invite_link') }}</h3>
          <button class="modal-close" type="button" @click="closeModal">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <div v-if="submitting" class="modal-loading">
            <div class="spinner-sm"></div>
            <span>{{ t('loading') }}</span>
          </div>
          <template v-else-if="inviteUrl">
            <div class="invite-url-box">
              <span class="invite-url-text">{{ inviteUrl }}</span>
              <button class="copy-url-btn" type="button" @click="copyInvite">
                <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor">
                  <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
                </svg>
                {{ inviteCopied ? t('copied') : t('copy') }}
              </button>
            </div>
            <div v-if="isAdmin" class="revoke-hint">{{ t('group_detail_revoke_hint') }}</div>
          </template>
          <div v-if="modalError" class="modal-error">{{ modalError }}</div>
        </div>
        <div class="modal-footer">
          <button v-if="isAdmin && inviteUrl" class="btn-danger-outline" type="button" @click="revokeInvite" :disabled="submitting">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor">
              <path d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"/>
            </svg>
            {{ t('group_detail_revoke_invite') }}
          </button>
          <button class="btn-cancel" type="button" @click="closeModal">{{ t('close') }}</button>
        </div>
      </div>
    </div>

    <!-- Add Participant -->
    <div v-if="modal === 'addParticipant'" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card">
        <div class="modal-header">
          <div class="modal-icon">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
              <path d="M15 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm-9-2V7H4v3H1v2h3v3h2v-3h3v-2H6zm9 4c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"/>
            </svg>
          </div>
          <h3>{{ t('group_detail_add_participants_title') }}</h3>
          <button class="modal-close" type="button" @click="closeModal">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="field">
            <label class="field-label">{{ t('groups_field_participants') }}</label>
            <textarea v-model="formData.participantsRaw" ref="firstInputRef" class="field-input field-textarea" rows="3" :placeholder="t('groups_field_participants_placeholder')"></textarea>
            <span class="field-hint">{{ t('groups_field_participants_hint') }}</span>
          </div>
          <div v-if="modalError" class="modal-error">{{ modalError }}</div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" type="button" @click="closeModal" :disabled="submitting">{{ t('cancel') }}</button>
          <button class="btn-confirm" type="button" @click="submitModal" :disabled="submitting || !formData.participantsRaw.trim()">
            {{ submitting ? t('saving') : t('group_detail_add') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Participant context menu (remove / promote / demote) -->
    <div v-if="modal === 'participantMenu' && selectedParticipant" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card modal-card-sm">
        <div class="modal-header">
          <div class="modal-icon">
            <div class="participant-avatar-placeholder-sm">
              {{ initials(selectedParticipant.DisplayName || participantPrimaryId(selectedParticipant)) }}
            </div>
          </div>
          <div>
            <h3 style="margin:0;font-size:15px">{{ selectedParticipant.DisplayName || formatPhone(participantPrimaryId(selectedParticipant)) }}</h3>
            <small style="color:#6b7280">{{ selectedParticipant.IsAdmin ? t('group_detail_admin') : t('group_detail_member') }}</small>
          </div>
          <button class="modal-close" type="button" @click="closeModal">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
          </button>
        </div>
        <div class="modal-body" style="padding:8px 16px 12px">
          <div v-if="modalError" class="modal-error" style="margin-bottom:8px">{{ modalError }}</div>
          <div class="participant-menu-actions">
            <button v-if="!selectedParticipant.IsAdmin" class="pmenu-btn pmenu-promote" type="button" @click="changeParticipantRole('promote')" :disabled="submitting">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M12 2l-5.5 9h11L12 2zm0 3.84L13.93 9h-3.87L12 5.84zM17.5 13c-2.49 0-4.5 2.01-4.5 4.5S15.01 22 17.5 22s4.5-2.01 4.5-4.5-2.01-4.5-4.5-4.5zm2.5 5h-2v2h-1v-2h-2v-1h2v-2h1v2h2v1zM11 13.5H2v8h9v-8zM10 20H3v-5.5h7V20z"/></svg>
              {{ t('group_detail_promote_admin') }}
            </button>
            <button v-else class="pmenu-btn pmenu-demote" type="button" @click="changeParticipantRole('demote')" :disabled="submitting">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M12 2l-5.5 9h11L12 2zm0 3.84L13.93 9h-3.87L12 5.84zM17.5 13c-2.49 0-4.5 2.01-4.5 4.5S15.01 22 17.5 22s4.5-2.01 4.5-4.5-2.01-4.5-4.5-4.5zm2.5 5h-5v-1h5v1zM11 13.5H2v8h9v-8zM10 20H3v-5.5h7V20z"/></svg>
              {{ t('group_detail_demote_admin') }}
            </button>
            <button class="pmenu-btn pmenu-remove" type="button" @click="removeParticipant" :disabled="submitting">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M14 8c0-2.21-1.79-4-4-4S6 5.79 6 8s1.79 4 4 4 4-1.79 4-4zm3 2v2h6v-2h-6zM2 18v2h16v-2c0-2.66-5.33-4-8-4s-8 1.34-8 4z"/></svg>
              {{ t('group_detail_remove_participant') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Join Requests -->
    <div v-if="modal === 'requests'" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card modal-card-lg">
        <div class="modal-header">
          <div class="modal-icon">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
            </svg>
          </div>
          <h3>{{ t('group_detail_join_requests') }}</h3>
          <button class="modal-close" type="button" @click="closeModal">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <div v-if="submitting" class="modal-loading">
            <div class="spinner-sm"></div>
          </div>
          <div v-else-if="pendingRequests.length === 0" class="empty-requests">
            <svg viewBox="0 0 24 24" width="36" height="36" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/></svg>
            <p>{{ t('group_detail_no_requests') }}</p>
          </div>
          <div v-else class="requests-list">
            <div v-for="req in pendingRequests" :key="req.JID" class="request-row">
              <div class="participant-avatar-placeholder-sm" style="width:38px;height:38px;font-size:14px">
                {{ initials(req.DisplayName || req.JID) }}
              </div>
              <div class="request-info">
                <span class="request-name">{{ req.DisplayName || formatPhone(req.JID) }}</span>
                <span class="request-date">{{ formatDate(req.RequestedAt) }}</span>
              </div>
              <div class="request-actions">
                <button class="req-approve-btn" type="button" @click="handleRequest(req.JID, 'approve')" :disabled="submitting">
                  <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
                  {{ t('group_detail_approve') }}
                </button>
                <button class="req-reject-btn" type="button" @click="handleRequest(req.JID, 'reject')" :disabled="submitting">
                  <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
                  {{ t('group_detail_reject') }}
                </button>
              </div>
            </div>
          </div>
          <div v-if="modalError" class="modal-error" style="margin-top:10px">{{ modalError }}</div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" type="button" @click="closeModal">{{ t('close') }}</button>
        </div>
      </div>
    </div>

    <!-- Leave Confirm -->
    <div v-if="modal === 'leave'" class="modal-overlay" @click.self="closeModal">
      <div class="modal-card modal-card-sm">
        <div class="modal-header">
          <div class="modal-icon modal-icon-danger">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
              <path d="M17 7l-1.41 1.41L18.17 11H8v2h10.17l-2.58 2.58L17 17l5-5zM4 5h8V3H4c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h8v-2H4V5z"/>
            </svg>
          </div>
          <h3>{{ t('group_detail_leave_group') }}</h3>
          <button class="modal-close" type="button" @click="closeModal">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor"><path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <p style="color:#374151;margin:0;line-height:1.5">{{ t('group_detail_confirm_leave_text') }}</p>
          <div v-if="modalError" class="modal-error" style="margin-top:10px">{{ modalError }}</div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" type="button" @click="closeModal" :disabled="submitting">{{ t('cancel') }}</button>
          <button class="btn-danger" type="button" @click="submitModal" :disabled="submitting">
            {{ submitting ? t('processing') : t('group_detail_leave_group') }}
          </button>
        </div>
      </div>
    </div>

  </div>
</template>

<script lang="ts">
import { computed, defineComponent, nextTick, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/services/api'
import { pushToast } from '@/services/toast'
import { useLocale } from '@/i18n'

interface Participant { JID: string; PhoneNumber: string; LID: string; IsAdmin: boolean; IsSuperAdmin: boolean; DisplayName: string }
interface JoinRequest { JID: string; DisplayName: string; RequestedAt: string }

const AVATAR_COLORS = ['#7C3AED','#2563EB','#059669','#D97706','#DC2626','#0891B2','#65A30D','#9333EA']

type ModalType = '' | 'name' | 'topic' | 'photo' | 'invite' | 'addParticipant' | 'participantMenu' | 'requests' | 'leave'

export default defineComponent({
  setup() {
    const route = useRoute()
    const router = useRouter()
    const token = route.params.token as string
    const groupid = route.params.id as string
    const { t, locale } = useLocale()

    const group = ref<any>({})
    const groupPicture = ref('')
    const participantPictures = ref<Record<string, string>>({})
    const messages = ref<any[]>([])
    const visibleMessages = ref<any[]>([])
    const loading = ref(false)
    const error = ref('')
    const showSearch = ref(false)
    const participantSearch = ref('')
    const messagesListRef = ref<HTMLElement | null>(null)
    const isLoadingMore = ref(false)
    const mySessionWid = ref('')
    const pendingRequests = ref<JoinRequest[]>([])
    const PAGE_SIZE = 50

    // Modal state
    const modal = ref<ModalType>('')
    const submitting = ref(false)
    const modalError = ref('')
    const firstInputRef = ref<HTMLElement | null>(null)
    const formData = ref({ name: '', topic: '', photoUrl: '', participantsRaw: '' })
    const inviteUrl = ref('')
    const inviteCopied = ref(false)
    const selectedParticipant = ref<Participant | null>(null)

    const totalMessages = computed(() => messages.value.length)
    const hasMoreMessages = computed(() => visibleMessages.value.length < messages.value.length)

    function expandIdentifier(value: string | null | undefined) {
      const trimmed = String(value || '').trim().toLowerCase()
      if (!trimmed) return []

      const compact = trimmed.replace(/\s+/g, '')
      const noPlus = compact.replace(/^\+/, '')
      const beforeAt = noPlus.split('@')[0] || noPlus
      const beforeDevice = beforeAt.split(':')[0] || beforeAt

      return Array.from(new Set([compact, noPlus, beforeAt, beforeDevice].filter(Boolean)))
    }

    function collectIdentifiers(...values: Array<string | null | undefined>) {
      const identifiers = new Set<string>()
      for (const value of values) {
        for (const identifier of expandIdentifier(value)) identifiers.add(identifier)
      }
      return Array.from(identifiers)
    }

    function participantPrimaryId(participant: Partial<Participant> | null | undefined) {
      return participant?.PhoneNumber || participant?.JID || participant?.LID || ''
    }

    function participantUniqueKey(participant: Partial<Participant> | null | undefined) {
      return participant?.JID || participant?.LID || participant?.PhoneNumber || ''
    }

    function participantIdentifiers(participant: Partial<Participant> | null | undefined) {
      return collectIdentifiers(participant?.PhoneNumber, participant?.JID, participant?.LID)
    }

    const sessionIdentifiers = computed(() => collectIdentifiers(mySessionWid.value))

    const currentParticipant = computed<Participant | null>(() => {
      const participants = group.value.Participants || []
      if (!participants.length || sessionIdentifiers.value.length === 0) return null

      return participants.find((p: Participant) =>
        participantIdentifiers(p).some(identifier => sessionIdentifiers.value.includes(identifier))
      ) || null
    })

    const isAdmin = computed(() => Boolean(currentParticipant.value?.IsAdmin || currentParticipant.value?.IsSuperAdmin))

    const filteredParticipants = computed(() => {
      const list = group.value.Participants || []
      if (!participantSearch.value) return list
      const q = participantSearch.value.toLowerCase()
      const queryIdentifiers = collectIdentifiers(participantSearch.value)
      return list.filter((p: Participant) =>
        p.DisplayName?.toLowerCase().includes(q) ||
        participantIdentifiers(p).some(identifier =>
          identifier.includes(q) || queryIdentifiers.some(queryIdentifier => identifier.includes(queryIdentifier))
        )
      )
    })

    const avatarStyle = computed(() => {
      const idx = groupid.split('').reduce((a, c) => a + c.charCodeAt(0), 0) % AVATAR_COLORS.length
      return { background: AVATAR_COLORS[idx] }
    })

    function initials(name: string) {
      if (!name) return '?'
      return name.split(' ').slice(0, 2).map(w => w[0]).join('').toUpperCase()
    }

    function formatPhone(phone: string) {
      if (!phone) return ''
      const raw = phone.trim()

      if (raw.includes('@lid')) return raw

      const base = raw
        .replace('@s.whatsapp.net', '')
        .split(':')[0]
        .replace(/^\+/, '')

      if (!/^\d+$/.test(base)) return raw

      if (base.startsWith('55') && (base.length === 12 || base.length === 13)) {
        const country = base.slice(0, 2)
        const ddd = base.slice(2, 4)
        const subscriber = base.slice(4)
        const splitIndex = subscriber.length === 8 ? 4 : 5
        return `+${country} ${ddd} ${subscriber.slice(0, splitIndex)}-${subscriber.slice(splitIndex)}`
      }

      return `+${base}`
    }

    function formatDate(dateStr: string) {
      if (!dateStr) return ''
      return new Date(dateStr).toLocaleDateString(locale.value, {
        day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit',
      })
    }

    function formatTime(timestamp: string) {
      if (!timestamp) return ''
      const date = new Date(timestamp)
      const now = new Date()
      const diff = now.getTime() - date.getTime()
      const oneDay = 86400000
      if (diff < oneDay && date.getDate() === now.getDate())
        return date.toLocaleTimeString(locale.value, { hour: '2-digit', minute: '2-digit' })
      if (diff < 2 * oneDay) return t('time_yesterday')
      if (diff < 7 * oneDay) return date.toLocaleDateString(locale.value, { weekday: 'short' })
      return date.toLocaleDateString(locale.value, { day: '2-digit', month: '2-digit' })
    }

    function buildPreview(msg: any): string {
      let text = msg.text || ''
      if (msg.attachment) {
        const mime = msg.attachment.mimetype || ''
        if (mime.startsWith('image/')) text = text ? `[IMG] ${text}` : `[IMG] ${t('media_image')}`
        else if (mime.startsWith('video/')) text = text ? `[VID] ${text}` : `[VID] ${t('media_video')}`
        else if (mime.startsWith('audio/') || msg.type === 'ptt') text = text ? `[AUD] ${text}` : `[AUD] ${t('media_audio')}`
        else text = text ? `[ARQ] ${text}` : `[ARQ] ${t('media_file')}`
      }
      if (!text && msg.inreply) text = `[RPL] ${t('media_reply')}`
      return text
    }

    function messagePreview(msg: any, maxLen = 120) {
      const text = msg.text || ''
      return text.length <= maxLen ? text : text.substring(0, maxLen) + '…'
    }

    function loadMoreMessages() {
      if (!hasMoreMessages.value || isLoadingMore.value) return
      isLoadingMore.value = true
      const next = messages.value.slice(visibleMessages.value.length, visibleMessages.value.length + PAGE_SIZE)
      visibleMessages.value = visibleMessages.value.concat(next)
      requestAnimationFrame(() => { isLoadingMore.value = false })
    }

    function onMessagesScroll() {
      const el = messagesListRef.value
      if (!el || isLoadingMore.value || !hasMoreMessages.value) return
      if (el.scrollHeight - (el.scrollTop + el.clientHeight) < 160) loadMoreMessages()
    }

    // Modal control
    async function openModal(type: ModalType) {
      modal.value = type
      modalError.value = ''
      submitting.value = false
      inviteUrl.value = ''
      inviteCopied.value = false

      if (type === 'name') formData.value.name = group.value.Name || ''
      if (type === 'topic') formData.value.topic = group.value.Topic || ''
      if (type === 'photo') formData.value.photoUrl = ''
      if (type === 'addParticipant') formData.value.participantsRaw = ''

      if (type === 'invite') {
        submitting.value = true
        try {
          const res = await api.get('/api/groups/invite', { params: { token, groupId: groupid } })
          inviteUrl.value = res.data?.url || res.data?.inviteUrl || ''
        } catch (err: any) {
          modalError.value = err?.response?.data?.result || err?.message || t('group_detail_error_invite')
        } finally {
          submitting.value = false
        }
      }

      if (type === 'requests') {
        submitting.value = true
        try {
          const res = await api.get('/api/groups/requests', { params: { token, groupId: groupid } })
          pendingRequests.value = res.data?.requests || []
        } catch (err: any) {
          modalError.value = err?.response?.data?.result || err?.message || t('group_detail_error_requests')
        } finally {
          submitting.value = false
        }
      }

      await nextTick()
      ;(firstInputRef.value as HTMLInputElement | null)?.focus()
    }

    function closeModal() {
      if (submitting.value) return
      modal.value = ''
      selectedParticipant.value = null
    }

    function openParticipantMenu(p: Participant) {
      selectedParticipant.value = p
      modalError.value = ''
      submitting.value = false
      modal.value = 'participantMenu'
    }

    async function submitModal() {
      if (submitting.value) return
      submitting.value = true
      modalError.value = ''

      try {
        if (modal.value === 'name') {
          await api.put('/api/groups/name', { token, groupId: groupid, name: formData.value.name.trim() })
          pushToast(t('group_detail_name_updated'), 'success')
          group.value = { ...group.value, Name: formData.value.name.trim() }
          modal.value = ''
        } else if (modal.value === 'topic') {
          await api.put('/api/groups/description', { token, groupId: groupid, topic: formData.value.topic })
          pushToast(t('group_detail_description_updated'), 'success')
          group.value = { ...group.value, Topic: formData.value.topic }
          modal.value = ''
        } else if (modal.value === 'photo') {
          await api.put('/api/groups/photo', { token, groupId: groupid, image_url: formData.value.photoUrl.trim() })
          pushToast(t('group_detail_photo_updated'), 'success')
          modal.value = ''
          await loadGroupPicture()
        } else if (modal.value === 'addParticipant') {
          const participants = formData.value.participantsRaw.split(',').map(p => p.trim()).filter(Boolean)
          await api.put('/api/groups/participants', { token, groupId: groupid, action: 'add', participants })
          pushToast(t('group_detail_participants_added'), 'success')
          modal.value = ''
          await load()
        } else if (modal.value === 'leave') {
          await api.post('/api/groups/leave', { token, groupId: groupid })
          pushToast(t('group_detail_leave_requested'), 'success')
          modal.value = ''
          router.push(`/server/${token}/groups`)
        }
      } catch (err: any) {
        modalError.value = err?.response?.data?.result || err?.message || t('error')
      } finally {
        submitting.value = false
      }
    }

    async function copyInvite() {
      try {
        await navigator.clipboard.writeText(inviteUrl.value)
        inviteCopied.value = true
        pushToast(t('group_detail_invite_copied'), 'success')
        setTimeout(() => { inviteCopied.value = false }, 2000)
      } catch { /* ignore */ }
    }

    async function revokeInvite() {
      submitting.value = true
      modalError.value = ''
      try {
        await api.delete('/api/groups/invite', { data: { token, groupId: groupid } })
        inviteUrl.value = ''
        pushToast(t('group_detail_invite_revoked'), 'success')
        modal.value = ''
      } catch (err: any) {
        modalError.value = err?.response?.data?.result || err?.message || t('group_detail_error_invite')
      } finally {
        submitting.value = false
      }
    }

    async function changeParticipantRole(action: 'promote' | 'demote') {
      if (!selectedParticipant.value || submitting.value) return
      const participantId = participantPrimaryId(selectedParticipant.value)
      if (!participantId) return
      submitting.value = true
      modalError.value = ''
      try {
        await api.put('/api/groups/participants', {
          token, groupId: groupid, action,
          participants: [participantId],
        })
        pushToast(action === 'promote' ? t('group_detail_promoted') : t('group_detail_demoted'), 'success')
        modal.value = ''
        await load()
      } catch (err: any) {
        modalError.value = err?.response?.data?.result || err?.message || t('error')
      } finally {
        submitting.value = false
      }
    }

    async function removeParticipant() {
      if (!selectedParticipant.value || submitting.value) return
      const participantId = participantPrimaryId(selectedParticipant.value)
      if (!participantId) return
      submitting.value = true
      modalError.value = ''
      try {
        await api.put('/api/groups/participants', {
          token, groupId: groupid, action: 'remove',
          participants: [participantId],
        })
        pushToast(t('group_detail_participant_removed'), 'success')
        modal.value = ''
        await load()
      } catch (err: any) {
        modalError.value = err?.response?.data?.result || err?.message || t('error')
      } finally {
        submitting.value = false
      }
    }

    async function handleRequest(jid: string, action: 'approve' | 'reject') {
      submitting.value = true
      modalError.value = ''
      try {
        await api.post('/api/groups/requests', { token, groupId: groupid, action, participants: [jid] })
        pendingRequests.value = pendingRequests.value.filter(r => r.JID !== jid)
        pushToast(action === 'approve' ? t('group_detail_request_approved') : t('group_detail_request_rejected'), 'success')
        if (action === 'approve') await load()
      } catch (err: any) {
        modalError.value = err?.response?.data?.result || err?.message || t('error')
      } finally {
        submitting.value = false
      }
    }

    async function load() {
      loading.value = true
      error.value = ''
      messages.value = []
      visibleMessages.value = []
      try {
        const res = await api.post('/api/groups/get', { token, groupId: groupid })
        group.value = res.data?.groupinfo || {}
        await loadGroupPicture()
        try {
          const serverRes = await api.post('/api/sessions/get', { token })
          const summary = serverRes.data?.server || {}
          mySessionWid.value = String(summary.wid || summary.Wid || serverRes.data?.wid || '').trim()
        } catch { /* ignore */ }
        await loadParticipantPictures()
        await loadMessages()
      } catch (err: any) {
        error.value = err?.response?.data?.result || err?.message || t('group_detail_error_load')
      } finally {
        loading.value = false
      }
    }

    async function loadGroupPicture() {
      try {
        const res = await api.post('/api/media/pictures/info', { token, chatId: groupid })
        if (res.data?.info?.url) groupPicture.value = res.data.info.url
      } catch { /* ignore */ }
    }

    async function loadParticipantPictures() {
      const list = (group.value.Participants || []).slice(0, 10)
      for (const p of list) {
        const id = participantPrimaryId(p)
        if (!id) continue
        try {
          const res = await api.post('/api/media/pictures/info', { token, chatId: id })
          if (res.data?.info?.url) participantPictures.value[id] = res.data.info.url
        } catch { /* ignore */ }
      }
    }

    async function loadMessages() {
      try {
        const res = await api.get('/api/messages', { params: { token } })
        const raw = res.data?.messages || []
        const grouped: any[] = []
        for (const msg of raw) {
          if (msg.chat?.id !== groupid) continue
          if (['unhandled','revoked','system'].includes(msg.type)) continue
          if (msg.debug?.reason === 'discard') continue
          const preview = buildPreview(msg)
          if (!preview) continue
          grouped.push({ ...msg, text: preview })
        }
        grouped.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
        messages.value = grouped
        visibleMessages.value = grouped.slice(0, PAGE_SIZE)
      } catch { /* optional */ }
    }

    onMounted(load)

    return {
      t, token, group, groupPicture, participantPictures,
      messages, visibleMessages, loading, error, showSearch, participantSearch,
      messagesListRef, isLoadingMore, totalMessages, hasMoreMessages,
      isAdmin, filteredParticipants, avatarStyle, pendingRequests,
      participantPrimaryId, participantUniqueKey,
      initials, formatPhone, formatDate, formatTime, messagePreview,
      onMessagesScroll, loadMoreMessages,
      modal, submitting, modalError, firstInputRef, formData,
      inviteUrl, inviteCopied, selectedParticipant,
      openModal, closeModal, openParticipantMenu, submitModal,
      copyInvite, revokeInvite, changeParticipantRole, removeParticipant, handleRequest,
    }
  },
})
</script>

<style scoped>
.group-detail-page {
  max-width: 1100px;
  margin: 0 auto;
}

.loading-state { text-align: center; padding: 80px 0; color: #6b7280; }

.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid #e5e7eb;
  border-top-color: var(--branding-primary, #7C3AED);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 14px;
}

@keyframes spin { to { transform: rotate(360deg); } }

.error-full { text-align: center; padding: 80px 0; color: #dc2626; }
.error-full svg { color: #fca5a5; margin-bottom: 14px; }
.error-full p { font-size: 15px; margin: 0 0 16px; }

.btn-back {
  display: inline-flex;
  align-items: center;
  padding: 10px 20px;
  background: var(--branding-primary, #7C3AED);
  color: white;
  border-radius: 10px;
  text-decoration: none;
  font-weight: 600;
  font-size: 14px;
}

.page-header { margin-bottom: 20px; }

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #334155;
  background: #f8fafc;
  border: 1px solid #dbe3ef;
  border-radius: 10px;
  padding: 6px 12px;
  font-size: 14px;
  font-weight: 600;
  text-decoration: none;
  transition: all 0.2s;
}

.back-link:hover { background: #eef2ff; border-color: #c7d2fe; color: #312e81; }

.group-layout {
  display: grid;
  grid-template-columns: 340px 1fr;
  gap: 20px;
  align-items: start;
}

/* Profile card */
.profile-card {
  background: white;
  border-radius: 16px;
  border: 1px solid #e5e7eb;
  padding: 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  margin-bottom: 14px;
  gap: 14px;
}

.profile-photo-wrap { position: relative; display: inline-block; }

.profile-photo {
  width: 96px; height: 96px;
  border-radius: 50%;
  object-fit: cover;
  border: 3px solid #e5e7eb;
}

.profile-photo-placeholder {
  width: 96px; height: 96px;
  border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  color: white; font-size: 32px; font-weight: 700;
}

.edit-photo-btn {
  position: absolute; bottom: 2px; right: 2px;
  width: 28px; height: 28px;
  border-radius: 50%;
  background: var(--branding-primary, #7C3AED);
  color: white; border: 2px solid white;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  transition: all 0.2s;
}

.edit-photo-btn:hover { filter: brightness(1.15); }

.profile-info { width: 100%; }

.profile-name-row {
  display: flex; align-items: center; justify-content: center;
  gap: 8px; margin-bottom: 4px;
}

.profile-name { font-size: 20px; font-weight: 700; color: #111827; margin: 0; }
.profile-meta { color: #6b7280; font-size: 13px; margin: 0; }

.edit-inline-btn {
  display: inline-flex; align-items: center; justify-content: center;
  width: 24px; height: 24px;
  border: none; background: #f3f4f6; border-radius: 6px;
  color: #6b7280; cursor: pointer; transition: all 0.2s; flex-shrink: 0;
}

.edit-inline-btn:hover { background: #e5e7eb; color: var(--branding-primary, #7C3AED); }

/* Info card */
.info-card {
  background: white; border-radius: 14px;
  border: 1px solid #e5e7eb; overflow: hidden; margin-bottom: 14px;
}

.card-header {
  display: flex; align-items: center; gap: 8px;
  padding: 12px 16px; background: #f9fafb;
  border-bottom: 1px solid #e5e7eb;
  font-size: 13px; font-weight: 600; color: #374151;
}

.card-header svg { color: var(--branding-primary, #7C3AED); flex-shrink: 0; }
.card-header span { flex: 1; }

.description-text { padding: 14px 16px 10px; margin: 0; color: #374151; white-space: pre-wrap; line-height: 1.5; font-size: 14px; }
.description-empty { padding: 14px 16px 10px; margin: 0; color: #9ca3af; font-style: italic; font-size: 14px; }
.description-meta { padding: 0 16px 12px; margin: 0; font-size: 11px; color: #9ca3af; }

/* Actions row */
.actions-row { display: flex; gap: 10px; margin-bottom: 14px; flex-wrap: wrap; }

.action-btn {
  flex: 1;
  min-width: 80px;
  display: flex; flex-direction: column; align-items: center; gap: 6px;
  padding: 14px 10px;
  background: white; border: 1px solid #e5e7eb; border-radius: 12px;
  color: var(--branding-primary, #7C3AED);
  font-size: 11px; font-weight: 600; cursor: pointer; transition: all 0.2s;
  position: relative;
}

.action-btn:hover { background: #f5f3ff; border-color: var(--branding-primary, #7C3AED); }

.req-badge {
  position: absolute; top: 8px; right: 8px;
  background: #ef4444; color: white;
  font-size: 10px; font-weight: 700;
  min-width: 16px; height: 16px;
  border-radius: 8px; padding: 0 4px;
  display: flex; align-items: center; justify-content: center;
}

/* Members */
.members-card .card-header { border-bottom: 1px solid #e5e7eb; }
.member-search-wrap { padding: 10px 14px 0; }
.member-search {
  width: 100%; padding: 8px 12px; border: 1px solid #e5e7eb;
  border-radius: 8px; font-size: 13px; box-sizing: border-box; transition: border-color 0.2s;
}
.member-search:focus { outline: none; border-color: var(--branding-primary, #7C3AED); }
.participants-list { max-height: 340px; overflow-y: auto; }

.participant-row {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 14px; border-bottom: 1px solid #f3f4f6;
}
.participant-row:last-child { border-bottom: none; }

.participant-avatar img { width: 38px; height: 38px; border-radius: 50%; object-fit: cover; }
.participant-avatar-placeholder {
  width: 38px; height: 38px; border-radius: 50%;
  background: #f3f4f6; display: flex; align-items: center; justify-content: center;
  color: #9ca3af; flex-shrink: 0;
}
.participant-avatar-placeholder-sm {
  width: 32px; height: 32px; border-radius: 50%;
  background: var(--branding-primary, #7C3AED);
  color: white; font-size: 12px; font-weight: 700;
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}

.participant-info { flex: 1; min-width: 0; }
.participant-name { font-size: 13px; font-weight: 500; color: #111827; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; display: block; }
.participant-phone { font-size: 11px; color: #6b7280; }
.participant-role { flex-shrink: 0; }
.participant-actions { flex-shrink: 0; }

.participant-action-btn {
  display: flex; align-items: center; justify-content: center;
  width: 26px; height: 26px; border-radius: 6px;
  background: #f3f4f6; border: none; color: #6b7280; cursor: pointer; transition: all 0.15s;
}
.participant-action-btn:hover { background: #e5e7eb; color: var(--branding-primary, #7C3AED); }

.role-badge { display: inline-block; padding: 2px 8px; border-radius: 20px; font-size: 10px; font-weight: 600; }
.role-owner { background: #f5f3ff; color: var(--branding-primary, #7C3AED); }
.role-admin { background: #dcfce7; color: #16a34a; }

/* Danger */
.danger-card { background: #fef2f2; border: 1px solid #fecaca; border-radius: 12px; overflow: hidden; }
.danger-btn {
  width: 100%; display: flex; align-items: center; justify-content: center; gap: 8px;
  padding: 14px; background: transparent; border: none; color: #dc2626;
  font-size: 14px; font-weight: 600; cursor: pointer; transition: background 0.2s;
}
.danger-btn:hover { background: #fee2e2; }

/* Messages col */
.messages-card { background: #0f172a; color: #e5e7eb; border-radius: 16px; padding: 18px; min-height: 500px; }
.messages-header { display: flex; justify-content: space-between; align-items: baseline; margin-bottom: 14px; }
.eyebrow { margin: 0; text-transform: uppercase; letter-spacing: 0.1em; font-size: 10px; color: #94a3b8; }
.messages-header h2 { margin: 0; font-size: 18px; }
.messages-header small { color: #94a3b8; font-size: 12px; }
.messages-list { display: flex; flex-direction: column; gap: 10px; max-height: calc(100vh - 240px); overflow-y: auto; padding-right: 4px; }
.message-item { display: grid; grid-template-columns: 38px 1fr; gap: 10px; padding: 10px 12px; border-radius: 12px; background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.06); }
.msg-avatar img { width: 38px; height: 38px; border-radius: 10px; object-fit: cover; }
.msg-avatar-placeholder { width: 38px; height: 38px; border-radius: 10px; background: #1f2937; display: flex; align-items: center; justify-content: center; color: #94a3b8; }
.msg-body { overflow: hidden; }
.msg-meta { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
.msg-sender { font-size: 12px; font-weight: 600; color: #fff; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.msg-time { font-size: 11px; color: #94a3b8; white-space: nowrap; flex-shrink: 0; }
.msg-text { font-size: 13px; color: #e2e8f0; line-height: 1.5; word-break: break-word; margin: 0; }
.load-more-area { text-align: center; padding: 10px 0; }
.load-more-info { color: #94a3b8; font-size: 12px; }
.load-more-info.muted { opacity: 0.6; }
.load-more-btn { background: #1e293b; color: #e2e8f0; border: none; padding: 8px 16px; border-radius: 8px; cursor: pointer; font-size: 13px; font-weight: 600; transition: background 0.2s; }
.load-more-btn:hover { background: #334155; }
.messages-empty { color: #94a3b8; text-align: center; padding: 30px 0; font-size: 14px; }

/* ===== MODALS ===== */
.modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.45);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000; padding: 20px;
}

.modal-card {
  background: white; border-radius: 18px;
  width: 100%; max-width: 460px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.2); overflow: hidden;
}

.modal-card-sm { max-width: 360px; }
.modal-card-lg { max-width: 540px; }

.modal-header {
  display: flex; align-items: center; gap: 12px;
  padding: 18px 20px 14px; border-bottom: 1px solid #f3f4f6;
}

.modal-icon {
  width: 38px; height: 38px; border-radius: 10px;
  background: #f5f3ff; color: var(--branding-primary, #7C3AED);
  display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}

.modal-icon-danger { background: #fef2f2; color: #dc2626; }

.modal-header h3 { flex: 1; font-size: 16px; font-weight: 700; color: #111827; margin: 0; }

.modal-close {
  display: flex; align-items: center; justify-content: center;
  width: 30px; height: 30px; background: none; border: none; border-radius: 8px;
  color: #9ca3af; cursor: pointer; transition: all 0.15s;
}
.modal-close:hover { background: #f3f4f6; color: #374151; }

.modal-body { padding: 18px 20px; display: flex; flex-direction: column; gap: 14px; }
.modal-footer {
  display: flex; align-items: center; justify-content: flex-end; gap: 10px;
  padding: 14px 20px 18px; border-top: 1px solid #f3f4f6;
}

.field { display: flex; flex-direction: column; gap: 6px; }
.field-label { font-size: 13px; font-weight: 600; color: #374151; }
.field-input {
  padding: 10px 14px; border: 2px solid #e5e7eb; border-radius: 10px;
  font-size: 14px; background: #f9fafb; transition: all 0.2s; font-family: inherit; resize: vertical;
}
.field-input:focus { outline: none; border-color: var(--branding-primary, #7C3AED); background: white; box-shadow: 0 0 0 4px rgba(124,58,237,0.08); }
.field-textarea { min-height: 80px; }
.field-hint { font-size: 11px; color: #9ca3af; }

.modal-error {
  display: flex; align-items: center; gap: 8px;
  padding: 10px 14px; background: #fef2f2; border: 1px solid #fecaca;
  border-radius: 8px; color: #dc2626; font-size: 13px;
}

.modal-loading { display: flex; align-items: center; gap: 12px; color: #6b7280; padding: 8px 0; }

.spinner-sm {
  width: 22px; height: 22px; border: 2px solid #e5e7eb;
  border-top-color: var(--branding-primary, #7C3AED);
  border-radius: 50%; animation: spin 0.8s linear infinite;
}

/* Invite modal */
.invite-url-box {
  display: flex; align-items: center; gap: 10px;
  background: #f9fafb; border: 1px solid #e5e7eb;
  border-radius: 10px; padding: 10px 14px;
}

.invite-url-text {
  flex: 1; font-size: 12px; color: #374151;
  word-break: break-all; font-family: monospace;
}

.copy-url-btn {
  flex-shrink: 0; display: inline-flex; align-items: center; gap: 5px;
  padding: 6px 12px; background: var(--branding-primary, #7C3AED);
  color: white; border: none; border-radius: 8px;
  font-size: 12px; font-weight: 600; cursor: pointer; transition: all 0.2s; white-space: nowrap;
}

.copy-url-btn:hover { filter: brightness(1.1); }
.revoke-hint { font-size: 12px; color: #9ca3af; }

/* Participant menu */
.participant-menu-actions { display: flex; flex-direction: column; gap: 4px; }

.pmenu-btn {
  display: flex; align-items: center; gap: 10px;
  width: 100%; padding: 12px 14px; border: none; border-radius: 10px;
  font-size: 14px; font-weight: 500; cursor: pointer; transition: all 0.15s; text-align: left;
}

.pmenu-promote { background: #f5f3ff; color: var(--branding-primary, #7C3AED); }
.pmenu-promote:hover { background: #ede9fe; }
.pmenu-demote { background: #fef9c3; color: #92400e; }
.pmenu-demote:hover { background: #fef08a; }
.pmenu-remove { background: #fef2f2; color: #dc2626; }
.pmenu-remove:hover { background: #fee2e2; }
.pmenu-btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* Join requests */
.empty-requests { text-align: center; padding: 24px 0; color: #9ca3af; }
.empty-requests svg { color: #d1d5db; margin-bottom: 10px; }
.empty-requests p { margin: 0; font-size: 14px; }

.requests-list { display: flex; flex-direction: column; gap: 8px; max-height: 360px; overflow-y: auto; }

.request-row {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 12px; background: #f9fafb; border-radius: 10px; border: 1px solid #e5e7eb;
}

.request-info { flex: 1; min-width: 0; }
.request-name { display: block; font-size: 13px; font-weight: 600; color: #111827; }
.request-date { font-size: 11px; color: #9ca3af; }

.request-actions { display: flex; gap: 6px; }

.req-approve-btn, .req-reject-btn {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 5px 10px; border: none; border-radius: 7px;
  font-size: 12px; font-weight: 600; cursor: pointer; transition: all 0.15s;
}

.req-approve-btn { background: #dcfce7; color: #16a34a; }
.req-approve-btn:hover:not(:disabled) { background: #bbf7d0; }
.req-reject-btn { background: #fef2f2; color: #dc2626; }
.req-reject-btn:hover:not(:disabled) { background: #fee2e2; }
.req-approve-btn:disabled, .req-reject-btn:disabled { opacity: 0.5; cursor: not-allowed; }

/* Buttons */
.btn-cancel {
  padding: 8px 18px; background: #f3f4f6; border: none; border-radius: 10px;
  font-size: 14px; font-weight: 600; color: #374151; cursor: pointer; transition: background 0.15s;
}
.btn-cancel:hover:not(:disabled) { background: #e5e7eb; }
.btn-cancel:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-confirm {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 8px 18px;
  background: linear-gradient(135deg, var(--branding-primary, #7C3AED), var(--branding-secondary, #5B21B6));
  color: white; border: none; border-radius: 10px;
  font-size: 14px; font-weight: 600; cursor: pointer; transition: all 0.2s;
}
.btn-confirm:hover:not(:disabled) { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(124,58,237,0.3); }
.btn-confirm:disabled { opacity: 0.5; cursor: not-allowed; transform: none; }

.btn-danger {
  padding: 8px 18px; background: #ef4444; border: none; border-radius: 10px;
  color: white; font-size: 14px; font-weight: 600; cursor: pointer; transition: all 0.2s;
}
.btn-danger:hover:not(:disabled) { background: #dc2626; }
.btn-danger:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-danger-outline {
  display: inline-flex; align-items: center; gap: 6px;
  padding: 8px 14px; background: white; border: 2px solid #ef4444;
  border-radius: 10px; color: #ef4444; font-size: 13px; font-weight: 600; cursor: pointer; transition: all 0.2s;
}
.btn-danger-outline:hover:not(:disabled) { background: #fef2f2; }
.btn-danger-outline:disabled { opacity: 0.5; cursor: not-allowed; }

.spin-icon { display: inline-block; animation: spin 0.8s linear infinite; }

html[data-theme='dark'] .back-link {
  background: rgba(15, 23, 42, 0.92);
  border-color: rgba(71, 85, 105, 0.28);
  color: #e2e8f0;
}

html[data-theme='dark'] .back-link:hover {
  background: rgba(30, 41, 59, 0.94);
  border-color: rgba(124, 58, 237, 0.3);
  color: #f8fafc;
}

html[data-theme='dark'] .profile-card,
html[data-theme='dark'] .info-card,
html[data-theme='dark'] .action-btn,
html[data-theme='dark'] .danger-card,
html[data-theme='dark'] .modal-card {
  background: rgba(15, 23, 42, 0.94);
  border-color: rgba(71, 85, 105, 0.28);
  box-shadow: 0 20px 40px rgba(2, 6, 23, 0.28);
}

html[data-theme='dark'] .profile-name,
html[data-theme='dark'] .participant-name,
html[data-theme='dark'] .card-header,
html[data-theme='dark'] .description-text,
html[data-theme='dark'] .request-name,
html[data-theme='dark'] .modal-header h3,
html[data-theme='dark'] .invite-url-text,
html[data-theme='dark'] .field-label {
  color: #f8fafc;
}

html[data-theme='dark'] .profile-meta,
html[data-theme='dark'] .description-meta,
html[data-theme='dark'] .description-empty,
html[data-theme='dark'] .participant-phone,
html[data-theme='dark'] .request-date,
html[data-theme='dark'] .load-more-info,
html[data-theme='dark'] .modal-loading,
html[data-theme='dark'] .revoke-hint,
html[data-theme='dark'] .field-hint {
  color: #94a3b8;
}

html[data-theme='dark'] .profile-photo {
  border-color: rgba(71, 85, 105, 0.42);
}

html[data-theme='dark'] .edit-photo-btn {
  border-color: rgba(15, 23, 42, 0.94);
}

html[data-theme='dark'] .edit-inline-btn,
html[data-theme='dark'] .participant-action-btn,
html[data-theme='dark'] .modal-close {
  background: rgba(30, 41, 59, 0.94);
  color: #94a3b8;
}

html[data-theme='dark'] .edit-inline-btn:hover,
html[data-theme='dark'] .participant-action-btn:hover,
html[data-theme='dark'] .modal-close:hover {
  background: rgba(51, 65, 85, 0.96);
  color: #f8fafc;
}

html[data-theme='dark'] .card-header,
html[data-theme='dark'] .modal-header,
html[data-theme='dark'] .modal-footer {
  background: rgba(15, 23, 42, 0.96);
  border-color: rgba(71, 85, 105, 0.26);
}

html[data-theme='dark'] .action-btn {
  color: #c4b5fd;
}

html[data-theme='dark'] .action-btn:hover {
  background: rgba(30, 41, 59, 0.96);
}

html[data-theme='dark'] .member-search,
html[data-theme='dark'] .field-input {
  background: rgba(11, 22, 40, 0.96);
  border-color: #334155;
  color: #e2e8f0;
}

html[data-theme='dark'] .member-search::placeholder,
html[data-theme='dark'] .field-input::placeholder {
  color: #64748b;
}

html[data-theme='dark'] .member-search:focus,
html[data-theme='dark'] .field-input:focus {
  background: rgba(16, 28, 49, 0.98);
  box-shadow: 0 0 0 4px rgba(124, 58, 237, 0.14);
}

html[data-theme='dark'] .participant-row {
  border-bottom-color: rgba(51, 65, 85, 0.3);
}

html[data-theme='dark'] .participant-avatar-placeholder {
  background: rgba(30, 41, 59, 0.96);
  color: #94a3b8;
}

html[data-theme='dark'] .role-owner {
  background: rgba(76, 29, 149, 0.34);
  color: #ddd6fe;
}

html[data-theme='dark'] .role-admin {
  background: rgba(20, 83, 45, 0.34);
  color: #bbf7d0;
}

html[data-theme='dark'] .danger-card {
  background: rgba(127, 29, 29, 0.16);
  border-color: rgba(239, 68, 68, 0.28);
}

html[data-theme='dark'] .danger-btn {
  color: #fca5a5;
}

html[data-theme='dark'] .danger-btn:hover {
  background: rgba(127, 29, 29, 0.24);
}

html[data-theme='dark'] .invite-url-box,
html[data-theme='dark'] .request-row {
  background: rgba(15, 23, 42, 0.92);
  border-color: rgba(71, 85, 105, 0.26);
}

html[data-theme='dark'] .copy-url-btn {
  box-shadow: none;
}

html[data-theme='dark'] .pmenu-promote {
  background: rgba(76, 29, 149, 0.28);
  color: #ddd6fe;
}

html[data-theme='dark'] .pmenu-promote:hover {
  background: rgba(76, 29, 149, 0.38);
}

html[data-theme='dark'] .pmenu-demote {
  background: rgba(120, 53, 15, 0.28);
  color: #fde68a;
}

html[data-theme='dark'] .pmenu-demote:hover {
  background: rgba(120, 53, 15, 0.38);
}

html[data-theme='dark'] .pmenu-remove {
  background: rgba(127, 29, 29, 0.22);
  color: #fca5a5;
}

html[data-theme='dark'] .pmenu-remove:hover {
  background: rgba(127, 29, 29, 0.32);
}

html[data-theme='dark'] .empty-requests svg {
  color: #475569;
}

html[data-theme='dark'] .btn-cancel {
  background: rgba(30, 41, 59, 0.94);
  color: #e2e8f0;
  border: 1px solid rgba(71, 85, 105, 0.28);
}

html[data-theme='dark'] .btn-cancel:hover:not(:disabled) {
  background: rgba(51, 65, 85, 0.96);
}

html[data-theme='dark'] .btn-danger-outline {
  background: rgba(15, 23, 42, 0.94);
  border-color: #ef4444;
  color: #fca5a5;
}

html[data-theme='dark'] .btn-danger-outline:hover:not(:disabled) {
  background: rgba(127, 29, 29, 0.2);
}

@media (max-width: 900px) {
  .group-layout { grid-template-columns: 1fr; }
  .messages-card { order: 2; }
}

@media (max-width: 600px) {
  .actions-row { flex-wrap: wrap; }
  .action-btn { min-width: calc(50% - 5px); }
  .modal-card { border-radius: 14px; }
}
</style>
