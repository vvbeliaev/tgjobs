<script lang="ts">
	import posthog from 'posthog-js';
	import {
		House,
		Search,
		CalendarDays,
		Settings,
		LogIn,
		PanelRight,
		Menu,
		Github
	} from 'lucide-svelte';
	import { afterNavigate, goto } from '$app/navigation';
	import { page } from '$app/state';

	import '$lib/shared/pb/pb-hook';
	import {
		ThemeLoad,
		PortalHost,
		Logo,
		Button,
		ThemeController,
		swipeable,
		uiStore,
		Sidebar
	} from '$lib';
	import { userStore } from '$lib/apps/user';
	import { jobsStore } from '$lib/apps/job';
	import favicon from '$lib/shared/assets/favicon.svg';

	import './layout.css';
	import PWA from './PWA.svelte';
	import Splash from './Splash.svelte';

	const nav = [{ label: 'Home', href: '/', icon: House }];

	let { children, data } = $props();
	const globalPromise = $derived(data.globalPromise);

	const user = $derived(userStore.user);

	// Posthog identify and set person
	$effect(() => {
		console.log(user);

		if (!user) return;

		posthog.identify(user.id, {
			email: user.email,
			name: user.name
		});
		posthog.capture('user_authenticated', {
			email: user.email,
			name: user.name
		});
	});

	// Global user load
	$effect(() => {
		globalPromise.then(({ userAuth, jobs }) => {
			if (userAuth) userStore.set(userAuth);
			if (jobs) jobsStore.set(jobs);
		});
	});

	// Real-time subscription
	$effect(() => {
		const userId = userStore.user?.id;
		if (!userId) return;
		userStore.subscribe();
		jobsStore.subscribe();
		return () => {
			userStore.unsubscribe();
			jobsStore.unsubscribe();
		};
	});

	afterNavigate(() => {
		uiStore.setSidebarOpen(false);
	});
</script>

<PWA />

<svelte:head
	><link rel="icon" href={favicon} />
	<link rel="icon" href={favicon} />
	<ThemeLoad />
</svelte:head>

{#snippet sidebarHeader({ expanded }: { expanded: boolean })}
	{#if expanded}
		<a href="/" class="flex items-center gap-2">
			<Logo />
		</a>
	{/if}
{/snippet}

{#snippet sidebarContent({ expanded }: { expanded: boolean })}
	<div class="shrink-0 space-y-1 px-2 pt-4">
		{#each nav as item}
			<Button
				class={[expanded ? 'justify-start' : '']}
				color={page.url.pathname === item.href ? 'primary' : 'neutral'}
				variant="ghost"
				block
				square={!expanded}
				href={item.href}
			>
				<item.icon class={expanded ? 'size-5' : 'size-6'} />
				{#if expanded}
					<span class="text-nowrap">{item.label}</span>
				{:else}
					<span class="sr-only">{item.label}</span>
				{/if}
			</Button>
		{/each}
	</div>
{/snippet}

{#snippet sidebarFooter({ expanded }: { expanded: boolean })}
	<!-- <div class="divider my-1"></div> -->

	{#if user && user.verified}
		<div class="mb-1 flex justify-center px-2">
			<!-- <button
				class={['btn justify-start btn-ghost', expanded ? 'btn-block' : 'btn-square']}
				onclick={() => uiStore.toggleFeedbackModal()}
			>
				<MessageSquare class={expanded ? 'size-5' : 'size-6'} />
				{#if expanded}
					Feedback
				{:else}
					<span class="sr-only">Feedback</span>
				{/if}
			</button> -->
		</div>
	{/if}

	<div class={['mb-3 flex flex-col border-base-300', expanded ? 'px-2' : 'items-center gap-3']}>
		<a
			href="https://github.com/vvbeliaev/tgjobs"
			target="_blank"
			rel="noopener noreferrer"
			class={['btn btn-ghost', expanded ? 'btn-block justify-start gap-2 px-4' : 'btn-square']}
			title="View on GitHub"
		>
			<Github class={expanded ? 'size-5' : 'size-8'} />
			{#if expanded}
				<span>GitHub</span>
			{/if}
		</a>

		<ThemeController {expanded} navStyle />
	</div>

	<div class="border-t border-base-300">
		{#if user && user.verified}
			<a
				href="/profile"
				class={[
					'flex items-center gap-3 p-2 transition-colors hover:bg-base-200',
					!expanded && 'justify-center'
				]}
				title={!expanded ? 'Settings' : ''}
			>
				{#if userStore.avatarUrl}
					<img src={userStore.avatarUrl} alt={user.name} class="size-10 rounded-full" />
				{:else}
					<div class="flex size-10 items-center justify-center rounded-full bg-base-300">
						{user.name?.at(0)?.toUpperCase() ?? 'U'}
					</div>
				{/if}
				{#if expanded}
					<div class="flex-1 overflow-hidden">
						<div class="truncate text-sm font-semibold">{user.name || '<No Name>'}</div>
						<div class="truncate text-xs opacity-60">{user.email}</div>
					</div>
					<Settings class="size-5 opacity-60" />
				{/if}
			</a>
		{:else}
			<a
				href="/auth"
				class={[
					'flex items-center gap-3 rounded-lg p-2 transition-colors hover:bg-base-300',
					!expanded && 'justify-center'
				]}
				title={!expanded ? 'Log in' : ''}
			>
				<div class="size-10 rounded-full bg-base-300"></div>
				{#if expanded}
					<div class="flex-1 overflow-hidden">
						<div class="truncate text-sm font-semibold">Log in</div>
					</div>
				{/if}
			</a>
		{/if}
	</div>
{/snippet}

{#await globalPromise}
	<Splash />
{:then}
	<div
		class="flex h-screen flex-col overflow-hidden bg-base-100 md:flex-row"
		use:swipeable={{
			isOpen: uiStore.sidebarOpen ?? false,
			direction: 'right',
			onOpen: () => uiStore.setSidebarOpen(true),
			onClose: () => uiStore.setSidebarOpen(false)
		}}
	>
		<!-- Sidebar -->
		<Sidebar
			open={uiStore.sidebarOpen ?? false}
			expanded={uiStore.sidebarExpanded ?? true}
			position="left"
			header={sidebarHeader}
			footer={sidebarFooter}
			onclose={() => uiStore.setSidebarOpen(false)}
			ontoggle={() => uiStore.toggleSidebarExpanded()}
		>
			{#snippet children({ expanded })}
				{@render sidebarContent({ expanded })}
			{/snippet}
		</Sidebar>

		<!-- Main Content -->
		<main class="mb-12 flex-1 overflow-y-auto md:mb-0">
			<div class="max-w-[1440px]">
				{@render children()}
			</div>
		</main>

		<!-- Mobile Dock -->
		<div class="dock dock-sm border-t border-base-300 md:hidden">
			<!-- Hidden for now -->
			<button onclick={() => uiStore.setSidebarOpen(true)}>
				<Menu class="size-5" />
				<span class="dock-label">Menu</span>
			</button>

			{#each nav as item}
				<a href={item.href} class:dock-active={page.url.pathname === item.href}>
					<item.icon class="size-5" />
					<span class="dock-label">{item.label}</span>
				</a>
			{/each}

			{#if user && user.verified}
				<a href="/profile" class:dock-active={page.url.pathname === '/profile'}>
					<Settings class="size-5" />
					<span class="dock-label">Profile</span>
				</a>
			{:else}
				<a href="/auth" class:dock-active={page.url.pathname === '/auth'}>
					<LogIn class="size-5" />
					<span class="dock-label">Log In</span>
				</a>
			{/if}
		</div>
	</div>
{/await}

<PortalHost />
