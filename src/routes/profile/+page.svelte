<script lang="ts">
	import { X } from 'lucide-svelte';
	import { goto } from '$app/navigation';

	import { pb, Button } from '$lib';
	import { userStore } from '$lib/apps/user';

	import ProfileCard from './ProfileCard.svelte';
	import LegalCard from './LegalCard.svelte';

	const user = $derived(userStore.user);

	function logout() {
		pb.authStore.clear();
		goto('/auth');
	}
</script>

<div class="mx-auto w-full max-w-7xl space-y-4 overflow-y-auto p-4 sm:space-y-0 sm:p-6 lg:p-4">
	<div class="mb-4 flex items-center justify-between sm:mb-6">
		<h1 class="mx-auto text-2xl font-bold sm:text-3xl">Profile & Settings</h1>
	</div>

	<div class={['mx-auto max-w-xl space-y-4']}>
		<ProfileCard />

		<!-- <LegalCard /> -->

		<Button class="mt-2 w-full" onclick={logout} variant="soft" color="error">
			<X class="mr-2 h-4 w-4" />
			Logout
		</Button>
	</div>
</div>
