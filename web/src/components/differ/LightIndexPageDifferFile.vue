<template>
  <div
    class="d2h-file-wrapper bg-white shadow rounded-lg my-4 z-0 relative overflow-y-hidden overflow-x-auto"
    :class="extraClasses"
  >
    <DiffHeader
      :diffs="diffs"
      :suggestions="null"
      :show-add="false"
      :show-suggestions="false"
      :is-added="false"
      :have-live-changes="false"
      :showing-suggestions-by-user="null"
      :can-ignore-file="false"
      :can-take-suggestions="false"
    />

    <div v-if="isReadyToDisplay">
      <div class="d2h-code-wrapper">
        <table
          class="d2h-diff-table leading-4 z-0"
          style="border-collapse: separate; border-spacing: 0"
        >
          <tbody
            v-for="(hunk, hunkIndex) in parsedHunks"
            :key="hunkIndex"
            :class="['d2h-diff-tbody d2h-file-diff z-0']"
          >
            <template
              v-for="block in highlightedBlocks(hunk.blocks, hunk.language)"
              :key="block.header"
            >
              <tr class="h-full overflow-hidden z-0">
                <td
                  class="d2h-code-linenumber d2h-info h-full sticky left-0 z-20 bg-white min-w-[80px]"
                ></td>
                <td class="bg-blue-50" />
                <td class="d2h-info h-full bg-blue-50 left-0 z-0 w-full">
                  <div class="flex items-center sticky left-0">
                    <div class="d2h-code-line d2h-info text-gray-500">
                      &nbsp;&nbsp;{{ block.header }}
                    </div>
                  </div>
                </td>
              </tr>

              <template v-for="(row, rowIndex) in block.lines" :key="rowIndex">
                <tr
                  :data-row-index="rowIndex"
                  class="z-0"
                  :data-preferred-name="diffs.preferred_name"
                  :data-line-oldnum="row.oldNumber"
                  :data-line-newnum="row.newNumber"
                >
                  <td
                    :class="[
                      'd2h-code-linenumber bg-white sticky left-0 z-20',
                      row.type === 'insert' ? 'border-r border-l border-green-500' : '',
                      row.type === 'delete' ? 'border-r border-l border-red-500' : '',
                    ]"
                  >
                    <div class="select-none text-gray-600 flex">
                      <div class="line-num">{{ row.oldNumber }}</div>
                      <div class="line-num">{{ row.newNumber }}</div>
                    </div>
                  </td>

                  <td
                    :class="[
                      row.type === 'insert' ? 'bg-green-50' : '',
                      row.type === 'delete' ? 'bg-red-50' : '',
                    ]"
                  ></td>

                  <td
                    :class="[
                      'code-row-wrapper relative z-10',
                      row.type === 'insert' ? 'bg-green-50' : '',
                      row.type === 'delete' ? 'bg-red-50' : '',
                    ]"
                  >
                    <div class="d2h-code-line relative z-0 px-4">
                      <span v-if="row.type === 'context'" class="d2h-code-line-prefix">&nbsp;</span>
                      <span v-else class="d2h-code-line-prefix">{{ row.prefix }}</span>
                      <span
                        v-if="row.content"
                        class="d2h-code-line-ctn whitespace-pre"
                        v-html="row.content"
                      />
                      <span v-else class="d2h-code-line-ctn whitespace-pre">{{
                        row.originalContent
                      }}</span>
                    </div>
                  </td>
                </tr>
              </template>
            </template>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType, reactive } from 'vue'
import { Block, HighlightedBlock } from './event'
import * as Diff2Html from 'diff2html'
import DiffHeader from './DiffHeader.vue'
import highlight from '../../highlight/highlight'

interface Data {
  parsedHunks: any
}

interface Pos {
  hunkIndex: number
  blockIndex: number
  rowIndex: number
}

export interface FileDiff {
  // unidiff.FileDiff
  orig_name: string
  new_name: string
  hunks: Array<string>
}

export default defineComponent({
  components: {
    DiffHeader,
  },
  emits: ['fileSelectedHunks', 'applyHunkedSuggestion', 'composeNewCommentAt'],
  data(): Data {
    return {
      parsedHunks: [], // Diff2Html objects
    }
  },
  props: {
    extraClasses: String,
    diffs: {
      type: Object as PropType<FileDiff>,
      required: true,
    },
  },
  watch: {
    diffs() {
      this.parse()
    },
  },
  created() {
    this.parse()
  },
  methods: {
    parse() {
      let res = []
      this.diffs.hunks.forEach((hunk) => {
        res = res.concat(
          Diff2Html.parse(hunk.patch, {
            matching: 'lines',
            outputFormat: 'line-by-line',
          })
        )
      })
      if (this.diffs.hunks.length > 0 && res.length !== this.diffs.hunks.length) {
        alert('Unexpected length in diff parsing?')
      }
      this.parsedHunks = res
      this.isReadyToDisplay = true
    },
    highlightedBlocks(input: Array<Block>, lang: string): Array<HighlightedBlock> {
      return highlight(input, lang)
    },
  },
})
</script>
