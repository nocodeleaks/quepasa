import axios from 'axios'

const api = axios.create({
  baseURL: '/',
  withCredentials: true,
  headers: { 'Accept': 'application/json' }
})

export default api
