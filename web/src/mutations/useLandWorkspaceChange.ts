import { gql, useMutation } from '@urql/vue'
import {
  LandWorkspaceChangeMutation,
  LandWorkspaceChangeMutationVariables,
} from './__generated__/useLandWorkspaceChange'
import { UpdateResolver } from '@urql/exchange-graphcache'
import { DeepMaybeRef } from '@vueuse/core'
import { Ref } from 'vue'
import { LandWorkspaceChangeInput } from '../__generated__/types'

export const LAND_WORKSPACE_CHANGE = gql<
  LandWorkspaceChangeMutation,
  DeepMaybeRef<LandWorkspaceChangeMutationVariables>
>`
  mutation LandWorkspaceChange($input: LandWorkspaceChangeInput!) {
    landWorkspaceChange(input: $input) {
      id
      upToDateWithTrunk
      draftDescription
    }
  }
`

export function useLandWorkspaceChange(): {
  mutating: Ref<boolean>
  landWorkspaceChange(input: DeepMaybeRef<LandWorkspaceChangeInput>): Promise<void>
} {
  const { executeMutation, fetching, error } = useMutation(LAND_WORKSPACE_CHANGE)

  return {
    mutating: fetching,
    landWorkspaceChange: async (input) => {
      const result = await executeMutation({ input })
      if (result.error) {
        throw result.error
      }
    },
  }
}

export const landWorkspaceChangeUpdateResolver: UpdateResolver<
  LandWorkspaceChangeMutation,
  LandWorkspaceChangeMutationVariables
> = (parent, args, cache) => {
  // Manually update cache if necessary
}
