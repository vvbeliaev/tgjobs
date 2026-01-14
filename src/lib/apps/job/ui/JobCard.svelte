<script lang="ts">
	import type { JobsResponse } from '$lib';
	import {
		Monitor,
		MapPin,
		Send,
		Archive,
		ArchiveRestore,
		ChevronDown,
		ChevronUp
	} from 'lucide-svelte';
	import { jobsStore } from '../jobs.svelte';
	import { slide } from 'svelte/transition';

	let { job }: { job: JobsResponse } = $props();
	let isExpanded = $state(false);

	function formatSalary(min?: number, max?: number, currency?: string) {
		if (!min && !max) return null;
		const curr = currency || 'USD';
		if (min && max) return `${min} - ${max} ${curr}`;
		if (min) return `from ${min} ${curr}`;
		if (max) return `up to ${max} ${curr}`;
		return null;
	}

	const salary = $derived(formatSalary(job.salaryMin, job.salaryMax, job.currency));

	function getTelegramUrl(url?: string, channelId?: string, messageId?: number) {
		let finalUrl = url;
		if (!finalUrl && channelId && messageId) {
			const cid = channelId.startsWith('-100') ? channelId.slice(4) : channelId;
			finalUrl = `https://t.me/c/${cid}/${messageId}`;
		}

		if (finalUrl?.includes('/c/-100')) {
			return finalUrl.replace('/c/-100', '/c/');
		}
		return finalUrl;
	}

	const tgUrl = $derived(getTelegramUrl(job.url, job.channelId, job.messageId));
</script>

<div
	class="group card border border-base-300 bg-base-100 shadow-sm transition-all hover:border-primary hover:shadow-md"
>
	<div class="card-body p-4 md:p-6">
		<div class="flex items-start justify-between gap-4">
			<div class="space-y-1">
				<h2 class="card-title text-lg transition-colors group-hover:text-primary md:text-xl">
					{job.title}
				</h2>
				{#if job.company}
					<p class="font-medium text-base-content/60">{job.company}</p>
				{/if}
			</div>
			<div class="flex items-center gap-2">
				{#if job.grade}
					<div class="badge badge-outline badge-md">{job.grade}</div>
				{/if}
				<button
					class="btn btn-circle btn-ghost btn-sm"
					onclick={() => jobsStore.toggleArchive(job.id)}
					title={job.archived ? 'Restore' : 'Archive'}
				>
					{#if job.archived}
						<ArchiveRestore size={18} class="text-success" />
					{:else}
						<Archive size={18} class="opacity-50 hover:opacity-100" />
					{/if}
				</button>
			</div>
		</div>

		<div class="mt-4 flex flex-wrap items-center gap-3">
			{#if job.isRemote}
				<div
					class="flex items-center gap-1.5 rounded-full bg-success/10 px-3 py-1 text-xs font-semibold text-success md:text-sm"
				>
					<Monitor size={14} />
					Remote
				</div>
			{:else if job.location}
				<div
					class="flex items-center gap-1.5 rounded-full bg-base-200 px-3 py-1 text-xs font-medium opacity-80 md:text-sm"
				>
					<MapPin size={14} />
					{job.location}
				</div>
			{/if}

			{#if salary}
				<div
					class="flex items-center gap-1.5 rounded-full bg-primary px-3 py-1 text-xs font-bold text-primary-content md:text-sm"
				>
					{salary}
				</div>
			{/if}

			{#if job.skills && Array.isArray(job.skills)}
				<div class="flex flex-wrap gap-2">
					{#each job.skills.slice(0, 5) as skill}
						<span class="text-xs opacity-50 before:mr-1 before:content-['#']">
							{skill}
						</span>
					{/each}
				</div>
			{/if}
		</div>

		{#if job.description}
			<p class="mt-4 line-clamp-2 text-sm leading-relaxed text-base-content/80 md:text-base">
				{job.description}
			</p>
		{/if}

		{#if isExpanded}
			<div transition:slide class="mt-4 space-y-4">
				<div class="rounded-lg bg-base-200 p-4">
					<h3 class="mb-2 text-xs font-bold tracking-wider uppercase opacity-50">Original Text</h3>
					<div
						class="max-h-148 overflow-y-auto text-sm leading-relaxed whitespace-pre-wrap opacity-90"
					>
						{job.originalText}
					</div>
				</div>
			</div>
		{/if}

		<div class="mt-6 flex items-center justify-between border-t border-base-200 pt-4">
			<button
				class="btn gap-1 opacity-50 btn-ghost btn-xs hover:opacity-100"
				onclick={() => (isExpanded = !isExpanded)}
			>
				{#if isExpanded}
					<ChevronUp size={14} />
					Show less
				{:else}
					<ChevronDown size={14} />
					Show more
				{/if}
			</button>

			<div class="flex items-center gap-4">
				<div class="text-xs opacity-40">
					{#if job.created}
						{new Date(job.created).toLocaleDateString()}
					{/if}
				</div>
				{#if tgUrl}
					<a
						href={tgUrl}
						target="_blank"
						rel="noopener noreferrer"
						class="btn gap-2 btn-sm btn-primary"
					>
						<Send size={14} />
						Apply
					</a>
				{/if}
			</div>
		</div>
	</div>
</div>
