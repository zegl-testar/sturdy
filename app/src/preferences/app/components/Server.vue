<template>
  <tr class="bg-white">
    <td class="pl-3 py-2 whitespace-nowrap">
      <div class="flex items-center">
        <div class="flex-shrink-0">
          <ServerStatus :server="server" @click="onServerStatusClick" />
        </div>
      </div>
    </td>
    <td
      v-for="field in fields"
      :key="field"
      class="px-3 py-2 whitespace-nowrap text-sm font-medium text-gray-900"
    >
      {{ field }}
    </td>
    <td class="px-3 py-2 whitespace-nowrap text-sm font-medium text-gray-900">
      <button
        type="button"
        :disabled="isImmutable"
        class="text-red-600 hover:text-red-900"
        :class="{ 'opacity-25 cursor-not-allowed': isImmutable }"
        @click.prevent="handleDelete"
      >
        Delete
      </button>
    </td>
    <td class="px-3 py-2 whitespace-nowrap text-sm font-medium text-gray-900">
      <button
        type="button"
        class="text-blue-600 hover:text-blue-900"
        :disabled="!server.isUp"
        :class="{ 'opacity-25 cursor-not-allowed': !server.isUp }"
        @click.prevent="handleOpen"
      >
        Open
      </button>
    </td>
  </tr>
</template>

<script>
import ServerStatus from './ServerStatus.vue'
import ipc from '../ipc'
import { remove as deleteServer, update as updateServer } from '../stores/servers'
import { toRefs } from 'vue'

const updateServerStatus = (cfg) => {
  ipc.isHostUp(cfg).then((isUp) => {
    updateServer({ ...cfg, isUp })
  })
}

export default {
  components: {
    ServerStatus,
  },
  props: {
    server: {
      type: Object,
      required: true,
    },
  },
  setup(props) {
    const { server } = toRefs(props)
    updateServerStatus(server.value)
  },
  computed: {
    fields() {
      if (this.server.host) {
        return [this.server.title, this.server.host]
      }
      return [this.server.title, new URL(this.server.webURL).host]
    },
    isImmutable() {
      return {
        Cloud: true,
        Development: true,
      }[this.server.title]
    },
  },
  methods: {
    handleOpen() {
      if (this.server.isUp) ipc.openHost(this.server)
    },
    handleDelete() {
      deleteServer(this.server)
      ipc.deleteHost(this.server)
    },
    onServerStatusClick() {
      updateServerStatus(this.server)
    },
  },
}
</script>
