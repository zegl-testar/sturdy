<template>
  <div v-if="!newWorkspace">
    When the codebase is set up on your local machine, a workspace will be created for you.
  </div>
  <div v-else>
    <RouterLinkButton
      :to="{
        name: 'workspaceHome',
        params: {
          codebaseSlug: newWorkspace.codebase.shortID,
          id: newWorkspace.id,
        },
      }"
    >
      Go to {{ newWorkspace.name }}
    </RouterLinkButton>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'
import { gql } from 'graphql-tag'
import { SetupUserViewsFragment } from '../../components/codebase/__generated__/SetupSturdyGoToWorkspaceStep'
import RouterLinkButton from '../../components/shared/RouterLinkButton.vue'

export const SETUP_USER_VIEWS = gql`
  fragment SetupUserViews on User {
    id
    views {
      workspace {
        id
        name
        codebase {
          id
          shortID
        }
      }
    }
  }
`

export default defineComponent({
  components: { RouterLinkButton },
  props: {
    codebase: {
      required: true,
      type: Object as PropType<{ id: string }>,
    },
    user: {
      type: Object as PropType<SetupUserViewsFragment | undefined>,
      default: undefined,
    },
  },
  computed: {
    newWorkspace(): SetupUserViewsFragment['views'][number]['workspace'] {
      return this.user?.views.filter((view) => view.workspace?.codebase.id === this.codebase.id)[0]
        ?.workspace
    },
  },
})
</script>
