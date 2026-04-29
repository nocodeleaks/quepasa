const en = {
  // App / session
  loading_session: 'Loading session...',
  logout: 'Logout',

  // Navbar
  nav_home: 'Home',
  nav_connect: 'Connect',
  nav_account: 'Account',
  nav_classic_ui: 'Classic UI',
  nav_api_docs: 'Swagger',
  nav_more: 'More',
  language_label: 'Language',

  // Home page
  home_title: 'Your Sessions',
  home_subtitle: 'Manage your WhatsApp connections',
  new_session: 'New Session',
  search_placeholder: 'Search...',
  search_sessions: 'Search sessions',
  card_view: 'Card view',
  table_view: 'Table view',

  // Stats
  total: 'Total',
  connected: 'Connected',
  disconnected: 'Disconnected',

  // Loading / error states
  loading_sessions: 'Loading sessions...',
  error_retry: 'Try again',
  no_sessions_configured: 'No session configured',
  no_sessions_description: 'Add your first WhatsApp session to start',
  connect_whatsapp: 'Connect WhatsApp',
  creating: 'Creating...',
  no_results: 'No sessions match your search',
  no_results_hint: 'Try a different query or clear the search.',

  // Table headers
  col_active: 'Active',
  col_phone: 'Phone',
  col_token: 'Token',
  col_dispatch: 'Dispatch',
  col_connection: 'Connection',
  col_actions: 'Actions',

  // Table tooltips
  session_active: 'Session active',
  session_not_verified: 'Session not verified',

  // Pagination
  showing: 'Showing',
  of: 'of',
  sessions_label: 'sessions',
  per_page: 'Per page:',
  page_indicator: 'Page {0} of {1}',
  prev_page: 'Previous page',
  next_page: 'Next page',

  // Actions (buttons / dropdown)
  open: 'Open',
  connect: 'Connect',
  remove: 'Remove',
  send: 'Send',
  messages: 'Messages',
  dispatching: 'Dispatching',
  rabbitmq: 'RabbitMQ',
  disconnect: 'Disconnect',
  debug: 'Debug',
  enable: 'Enable',
  disable: 'Disable',
  send_message: 'Send Message',

  // Feature toggles (dropdown labels)
  groups: 'Groups',
  broadcasts: 'Broadcasts',
  read_receipts: 'Read Receipts',
  calls: 'Calls',

  // Tri-state tooltips  — {0} = feature name
  tristate_on: '{0}: ON (forced)',
  tristate_off: '{0}: OFF (forced)',
  tristate_default: '{0}: System default',

  // Uptime
  uptime: 'Uptime',

  // Session detail page
  back: 'Back',
  loading_session_detail: 'Loading session...',
  not_connected: 'Not connected',
  status: 'Status',
  unknown: 'Unknown',

  // Confirm dialogs
  confirm_remove: 'Do you really want to REMOVE this session? This action cannot be undone.',
  confirm_disconnect: 'Do you really want to disconnect this session?',

  // Toast messages
  token_copied: 'Token copied!',
  session_updated: 'Session updated',
  error_update_session: 'Error updating session',
  session_disconnected: 'Session disconnected',
  error_disconnect: 'Error disconnecting',
  debug_updated: 'Debug updated',
  error_update_debug: 'Error updating debug',
  groups_updated: 'Groups updated',
  error_update_groups: 'Error updating groups',
  broadcasts_updated: 'Broadcasts updated',
  error_update_broadcasts: 'Error updating broadcasts',
  read_receipts_updated: 'Read receipts updated',
  error_update_read_receipts: 'Error updating read receipts',
  calls_updated: 'Calls updated',
  error_update_calls: 'Error updating calls',
  session_removed: 'Session removed',
  error_remove_session: 'Error removing session',
  session_created: 'Session created successfully!',
  error_create_session: 'Error creating session',
  error_token_not_received: 'Token not received from server',
  error_load_sessions: 'Error loading sessions',
  error_search_sessions: 'Error searching sessions',
}

export default en
export type Messages = typeof en
