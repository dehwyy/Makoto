<script lang="ts">
	import { CreateItem } from '$lib/api/fetches'
	import { Items } from '$lib/stores/items-store'
	import { Tags, type TagInitial } from '$lib/stores/tags-store'
	import { Button } from 'makoto-ui-svelte'
	import ItemEditor from '../item-editor.svelte'

	let isEdit = false
	const onFinalButtonClick = async (payload: {
		key: string
		value: string
		extra: string
		tags: string[]
	}) => {
		const { key, value, extra, tags } = payload
		// Update $Store
		Items.Add({
			id: 0,
			key,
			value,
			extra,
			tags
		})

		// Request 2 backend
		const response = await CreateItem({ key, value, extra, tags })

		Items.SetIdToItemAfterCreate(response.data?.itemId ?? 0)
	}
</script>

<ItemEditor isCreateItemMode={true} finalButtonText="Add" bind:isEdit {onFinalButtonClick}>
	<div class="w-full md:w-1/2">
		<Button onClick={() => (isEdit = !isEdit)}>
			<span class="text-lg font-ContentT font-[600]">Add new word</span></Button>
	</div>
</ItemEditor>
