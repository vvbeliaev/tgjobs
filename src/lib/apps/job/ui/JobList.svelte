<script lang="ts">
	import { jobsStore } from '../jobs.svelte';
	import JobCard from './JobCard.svelte';
	import JobFilters from './JobFilters.svelte';
	import { SearchX, Inbox } from 'lucide-svelte';

	const displayJobs = $derived(jobsStore.filteredJobs.length > 0 ? jobsStore.filteredJobs : []);
</script>

<div class="mx-auto max-w-5xl px-4 py-8">
	<JobFilters />

	{#if displayJobs.length === 0}
		<div class="flex flex-col items-center justify-center py-20 text-center opacity-40">
			<div class="mb-6 rounded-full bg-base-200 p-6">
				<SearchX size={48} strokeWidth={1} />
			</div>
			<h3 class="text-xl font-bold">No jobs found</h3>
			<p class="mt-2 max-w-xs text-sm">
				We couldn't find any jobs matching your current filters. Try adjusting your search or
				filters.
			</p>
			<button
				class="btn mt-6 text-primary btn-ghost"
				onclick={() => {
					jobsStore.search = '';
					jobsStore.filterGrade = '';
					jobsStore.filterRemote = null;
					jobsStore.showArchived = false;
				}}
			>
				Reset all filters
			</button>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-4">
			{#each displayJobs as job (job.id)}
				<JobCard {job} />
			{/each}
		</div>

		<div class="mt-8 flex items-center justify-center gap-2 text-sm opacity-30">
			<Inbox size={14} />
			Showing {displayJobs.length}
			{displayJobs.length === 1 ? 'job' : 'jobs'}
		</div>
	{/if}
</div>
