<template>
  <div class="flex -space-x-1">
    <Avatar
      v-for="author in inRangeAuthors"
      :key="author.id || author.user_id"
      :author="author"
      size="8"
      class="ring-2 ring-white block"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue'
import { Author } from '../differ/event'
import Avatar from './Avatar.vue'

export default defineComponent({
  components: { Avatar },
  data() {
    return {}
  },
  props: {
    authors: {
      type: Object as PropType<Array<Author>>,
    },
    max: {
      type: Number,
      default: 5,
    },
  },
  computed: {
    inRangeAuthors() {
      if (this.authors.length <= this.max) {
        return this.authors
      }
      return this.authors.slice(0, this.max)
    },
  },
})
</script>
