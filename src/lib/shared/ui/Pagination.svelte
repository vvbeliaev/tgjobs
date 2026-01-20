<script lang="ts">
	import { ChevronLeft, ChevronRight } from 'lucide-svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';

	interface Props {
		currentPage: number;
		totalPages: number;
		onPageChange?: (page: number) => void;
	}

	let { currentPage, totalPages, onPageChange }: Props = $props();

	function handlePageChange(newPage: number) {
		if (newPage < 1 || newPage > totalPages || newPage === currentPage) return;

		if (onPageChange) {
			onPageChange(newPage);
		} else {
			const url = new URL(page.url);
			url.searchParams.set('page', newPage.toString());
			goto(url.toString(), { keepFocus: true, noScroll: true });
		}
	}

	const pages = $derived.by(() => {
		const result: (number | string)[] = [];
		const delta = 2;

		for (let i = 1; i <= totalPages; i++) {
			if (i === 1 || i === totalPages || (i >= currentPage - delta && i <= currentPage + delta)) {
				result.push(i);
			} else if (result[result.length - 1] !== '...') {
				result.push('...');
			}
		}
		return result;
	});
</script>

{#if totalPages > 1}
	<div class="flex items-center justify-center gap-2 py-8">
		<div class="join">
			<button
				class="btn join-item btn-sm"
				disabled={currentPage <= 1}
				onclick={() => handlePageChange(currentPage - 1)}
				aria-label="Previous page"
			>
				<ChevronLeft size={18} />
			</button>

			{#each pages as p}
				{#if p === '...'}
					<button class="btn btn-disabled join-item btn-sm">...</button>
				{:else}
					<button
						class="btn join-item btn-sm {currentPage === p ? 'btn-primary' : 'btn-ghost'}"
						onclick={() => handlePageChange(p as number)}
					>
						{p}
					</button>
				{/if}
			{/each}

			<button
				class="btn join-item btn-sm"
				disabled={currentPage >= totalPages}
				onclick={() => handlePageChange(currentPage + 1)}
				aria-label="Next page"
			>
				<ChevronRight size={18} />
			</button>
		</div>
	</div>
{/if}
