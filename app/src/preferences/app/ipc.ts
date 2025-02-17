import { HostConfig } from './shims/ipc'

const listHosts = (): Promise<HostConfig[]> => {
  return window.ipc.listHosts()
}

const addHostConfig = async (cfg: HostConfig): Promise<void> => {
  window.ipc.addHostConfig(cfg)
}

const isHostUp = async (cfg: HostConfig): Promise<boolean> => {
  return window.ipc.isHostUp(cfg)
}

const openHost = async (cfg: HostConfig): Promise<void> => {
  window.ipc.openHost(cfg)
}

const deleteHost = async (cfg: HostConfig): Promise<void> => {
  window.ipc.deleteHostConfig(cfg)
}

export default {
  listHosts,
  addHostConfig,
  isHostUp,
  openHost,
  deleteHost,
}
