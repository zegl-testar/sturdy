<template>
  <VerticalNavigation :navigation="navigation" />
</template>

<script lang="ts">
import VerticalNavigation from '../../../organisms/VerticalNavigation.vue'
import { defineComponent } from 'vue'
import { useRoute } from 'vue-router'

export default defineComponent({
  components: { VerticalNavigation },
  setup() {
    let route = useRoute()

    const navigation = [
      {
        name: 'Settings',
        linkName: 'codebaseSettings',
        current: route.name === 'codebaseSettings',
      },
      {
        name: 'Team',
        linkName: 'codebaseSettingsTeam',
        current: route.name === 'codebaseSettingsTeam',
      },
      {
        name: 'Integrations',
        linkName: 'codebaseSettingsIntegrations',
        current: !!(
          route.name &&
          [
            'codebaseSettingsIntegrations',
            'codebaseSettingsAddBuildkite',
            'codebaseSettingsEditBuildkite',
          ].includes(route.name.toString())
        ),
      },
      {
        name: 'Security & ACLs',
        linkName: 'codebaseSettingsAcls',
        current: route.name === 'codebaseSettingsAcls',
      },
      {
        name: 'Restore Workspaces',
        linkName: 'codebaseSettingsWorkspaces',
        current: route.name === 'codebaseSettingsWorkspaces',
      },
    ]

    return {
      navigation,
    }
  },
})
</script>
