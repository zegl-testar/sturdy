<template>
  <section :id="title.toLowerCase()">
    <div class="relative space-y-20 sm:space-y-28">
      <div
        class="border-t border-slate-700/30 block absolute top-0 inset-x-0 h-[37.5rem] bg-gradient-to-b from-[#0c1322] top-0 xl:top-8"
      ></div>
      <div class="relative pt-16 max-w-6xl px-6 sm:px-6 md:px-8 mx-auto">
        <div
          class="flex flex-col items-center md:flex-row md:space-x-4 md:space-y-0 lg:gap-8 space-y-6"
        >
          <div :class="biggerRightSide ? 'md:w-1/3' : 'md:w-1/2'">
            <div class="w-24">
              <slot name="icon"></slot>
            </div>
            <h2 :class="['mt-4', 'font-semibold', 'text-' + accent]">{{ title }}</h2>
            <p class="mt-4 text-3xl sm:text-4xl font-extrabold tracking-tight text-slate-50">
              {{ subtitle }}
            </p>
            <div class="mt-4 max-w-3xl space-y-8">
              <slot name="left"></slot>
              <router-link
                v-if="link"
                class="hover:text-slate-300 hover:border-slate-300 hover:bg-slate-900/10 bg-slate-600/80 text-slate-200 border-transparent border focus:outline-none focus:ring-2 focus:ring-slate-400 focus:ring-offset-2 focus:ring-offset-slate-50 focus:outline-none focus:ring-2 group inline-flex items-center h-9 rounded text-sm font-semibold whitespace-nowrap px-3"
                :to="{ name: link }"
              >
                Learn more
                <ChevronRightIcon class="h-4 w-4 ml-1" />
              </router-link>
            </div>
            <div class="mt-12 max-w-3xl space-y-6">
              <slot name="bottomLeft"></slot>
            </div>
          </div>
          <div :class="biggerRightSide ? 'md:w-2/3' : 'md:w-1/2'">
            <slot name="right"></slot>
          </div>
        </div>
        <div>
          <slot name="bottom"></slot>
        </div>
      </div>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { ChevronRightIcon } from '@heroicons/vue/outline'

export default defineComponent({
  components: { ChevronRightIcon },
  props: {
    title: {
      type: String,
      required: true,
    },
    subtitle: {
      type: String,
      required: true,
    },
    link: {
      type: String,
      required: false,
    },
    accent: {
      type: String,
      required: true,
    },
    biggerRightSide: {
      type: Boolean,
      required: false,
    },
  },
})
</script>

<style scoped></style>
