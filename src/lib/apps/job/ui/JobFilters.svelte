<script lang="ts">
	import { jobsStore } from '../jobs.svelte';
	import { Search, X, Filter, Globe, Briefcase, Archive } from 'lucide-svelte';

	let search = $state('');

	$effect(() => {
		jobsStore.search = search;
	});
</script>

<div class="mb-8 space-y-4">
	<div class="flex flex-col gap-4 md:flex-row md:items-center">
		<!-- Search Bar -->
		<div class="relative flex-1">
			<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4 opacity-50">
				<Search size={20} />
			</div>
			<input
				type="text"
				placeholder="Search jobs, companies, skills..."
				class="input-bordered input h-12 w-full pl-12 transition-all focus:border-primary focus:ring-2 focus:ring-primary/20"
				bind:value={search}
			/>
			{#if search}
				<button
					class="btn absolute top-1/2 right-2 btn-circle -translate-y-1/2 btn-ghost btn-sm"
					onclick={() => (search = '')}
				>
					<X size={16} />
				</button>
			{/if}
		</div>

		<!-- Filters -->
		<div class="flex flex-wrap items-center gap-2">
			<div class="join w-full md:w-auto">
				<div class="join-item flex items-center bg-base-200 px-3 opacity-60">
					<Briefcase size={18} />
				</div>
				<select
					class="select-bordered select join-item min-w-[140px] select-md"
					bind:value={jobsStore.filterGrade}
				>
					<option value="">All Grades</option>
					<option value="Junior">Junior</option>
					<option value="Middle">Middle</option>
					<option value="Senior">Senior</option>
					<option value="Lead">Lead</option>
				</select>
			</div>

			<div
				class="flex h-12 items-center gap-3 rounded-lg border border-base-300 bg-base-100 px-4 transition-colors hover:bg-base-200"
			>
				<Globe size={18} class="opacity-60" />
				<span class="text-sm font-medium">Remote</span>
				<input
					type="checkbox"
					class="toggle toggle-primary toggle-sm"
					checked={jobsStore.filterRemote === true}
					onchange={(e) => (jobsStore.filterRemote = e.currentTarget.checked ? true : null)}
				/>
			</div>

			<div
				class="flex h-12 items-center gap-3 rounded-lg border border-base-300 bg-base-100 px-4 transition-colors hover:bg-base-200"
			>
				<Archive size={18} class="opacity-60" />
				<span class="text-sm font-medium">Archived</span>
				<input
					type="checkbox"
					class="toggle toggle-secondary toggle-sm"
					bind:checked={jobsStore.showArchived}
				/>
			</div>

			{#if jobsStore.search || jobsStore.filterGrade || jobsStore.filterRemote !== null || jobsStore.showArchived}
				<button
					class="btn gap-2 text-error btn-ghost btn-sm"
					onclick={() => {
						search = '';
						jobsStore.filterGrade = '';
						jobsStore.filterRemote = null;
						jobsStore.showArchived = false;
					}}
				>
					<X size={16} />
					Clear
				</button>
			{/if}
		</div>
	</div>
</div>
