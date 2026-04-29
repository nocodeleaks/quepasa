import type { Messages } from './en'

const pt: Messages = {
  // App / session
  loading_session: 'Carregando sessão...',
  logout: 'Sair',

  // Navbar
  nav_home: 'Início',
  nav_connect: 'Conectar',
  nav_account: 'Conta',
  nav_classic_ui: 'UI Clássica',
  nav_api_docs: 'Swagger',
  nav_more: 'Mais',
  language_label: 'Idioma',

  // Home page
  home_title: 'Suas Sessões',
  home_subtitle: 'Gerencie suas conexões WhatsApp',
  new_session: 'Nova Sessão',
  search_placeholder: 'Buscar...',
  search_sessions: 'Buscar sessões',
  card_view: 'Visualização em cartões',
  table_view: 'Visualização em tabela',

  // Stats
  total: 'Total',
  connected: 'Conectados',
  disconnected: 'Desconectados',

  // Loading / error states
  loading_sessions: 'Carregando sessões...',
  error_retry: 'Tentar novamente',
  no_sessions_configured: 'Nenhuma sessão configurada',
  no_sessions_description: 'Adicione sua primeira sessão WhatsApp para começar',
  connect_whatsapp: 'Conectar WhatsApp',
  creating: 'Criando...',
  no_results: 'Nenhuma sessão encontrada',
  no_results_hint: 'Tente uma busca diferente ou limpe o filtro.',

  // Table headers
  col_active: 'Ativo',
  col_phone: 'Telefone',
  col_token: 'Token',
  col_dispatch: 'Dispatch',
  col_connection: 'Conexão',
  col_actions: 'Ações',

  // Table tooltips
  session_active: 'Sessão ativa',
  session_not_verified: 'Sessão não verificada',

  // Pagination
  showing: 'Exibindo',
  of: 'de',
  sessions_label: 'sessões',
  per_page: 'Por página:',
  page_indicator: 'Página {0} de {1}',
  prev_page: 'Página anterior',
  next_page: 'Próxima página',

  // Actions (buttons / dropdown)
  open: 'Abrir',
  connect: 'Conectar',
  remove: 'Remover',
  send: 'Enviar',
  messages: 'Mensagens',
  dispatching: 'Despachos',
  rabbitmq: 'RabbitMQ',
  disconnect: 'Desconectar',
  debug: 'Debug',
  enable: 'Ativar',
  disable: 'Desativar',
  send_message: 'Enviar Mensagem',

  // Feature toggles (dropdown labels)
  groups: 'Grupos',
  broadcasts: 'Broadcasts',
  read_receipts: 'Confirmações de leitura',
  calls: 'Ligações',

  // Tri-state tooltips  — {0} = feature name
  tristate_on: '{0}: ATIVADO (forçado)',
  tristate_off: '{0}: DESATIVADO (forçado)',
  tristate_default: '{0}: Padrão do sistema',

  // Uptime
  uptime: 'Uptime',

  // Session detail page
  back: 'Voltar',
  loading_session_detail: 'Carregando sessão...',
  not_connected: 'Não conectado',
  status: 'Status',
  unknown: 'Desconhecido',

  // Confirm dialogs
  confirm_remove: 'Deseja realmente REMOVER esta sessão? Esta ação não pode ser desfeita.',
  confirm_disconnect: 'Deseja realmente desconectar esta sessão?',

  // Toast messages
  token_copied: 'Token copiado!',
  session_updated: 'Sessão atualizada',
  error_update_session: 'Erro ao alterar sessão',
  session_disconnected: 'Sessão desconectada',
  error_disconnect: 'Erro ao desconectar',
  debug_updated: 'Debug atualizado',
  error_update_debug: 'Erro ao alterar debug',
  groups_updated: 'Grupos atualizado',
  error_update_groups: 'Erro ao alterar grupos',
  broadcasts_updated: 'Broadcasts atualizado',
  error_update_broadcasts: 'Erro ao alterar broadcasts',
  read_receipts_updated: 'Confirmações de leitura atualizado',
  error_update_read_receipts: 'Erro ao alterar confirmações',
  calls_updated: 'Ligações atualizado',
  error_update_calls: 'Erro ao alterar ligações',
  session_removed: 'Sessão removida',
  error_remove_session: 'Erro ao remover sessão',
  session_created: 'Sessão criada com sucesso!',
  error_create_session: 'Erro ao criar sessão',
  error_token_not_received: 'Token não recebido do servidor',
  error_load_sessions: 'Erro ao carregar sessões',
  error_search_sessions: 'Erro ao buscar sessões',
}

export default pt
