<script lang="ts">
	import { userStore } from '$lib/apps/user';
	import JobList from '$lib/apps/job/ui/JobList.svelte';

	const user = $derived(userStore.user);
</script>

<div class="container mx-auto">
	{#if user}
		<JobList />
	{:else}
		<div class="flex min-h-[60vh] flex-col items-center justify-center p-4 text-center">
			<h1 class="mb-4 text-3xl font-bold">Finding your next job...</h1>
			<p class="mb-8 max-w-md text-xl opacity-70">
				Sign in to start tracking job offers from your Telegram channels automatically.
			</p>
			<a href="/auth" class="btn btn-lg btn-primary">Get Started</a>
		</div>
	{/if}
</div>
