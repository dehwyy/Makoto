<script lang="ts">
	// header's icons
	import user from '$lib/assets/logout.svg'
	//
	import { onMount } from 'svelte'
	import { Username } from '$lib/stores/user-store'

	$: isMounted = false
	onMount(() => {
		isMounted = true
	})
</script>

<header class="header">
	{#if $Username}
		<div
			class={`${
				isMounted ? 'opacity-100' : 'opacity-[1%]'
			} transition-all duration-300 delay-150 user_wrapper`}>
			<a href={`http://localhost:3000`}>
				<p class="user_item">{$Username}</p>
			</a>
		</div>
	{:else}<a href={`http://localhost:3001`}>
			<div class="icon_wrapper">
				<img class="icon_item" src={user} alt="logo" />
			</div>
		</a>
	{/if}
</header>

<style lang="scss">
	.header {
		@apply fixed top-0 left-0 right-0;
	}
	.user {
		&_wrapper {
			@apply py-5 px-10 mx-auto w-full flex justify-end;
		}
		&_item {
			@apply font-Jua text-2xl select-none cursor-pointer font-[700];
		}
	}
	.icon {
		&_wrapper {
			@apply cursor-pointer py-5 px-10 h-[40px] w-full mx-auto flex justify-end;
		}
		&_item {
			@apply h-[40px] w-[40px];
		}
	}
</style>
