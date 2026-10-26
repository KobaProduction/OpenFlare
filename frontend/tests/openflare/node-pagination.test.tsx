import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react';
import { NextIntlClientProvider } from 'next-intl';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';

import { NodesPageClient } from '@/app/(main)/nodes/page-client';
import { NodeDetailPageClient } from '@/app/(main)/nodes/detail/page-client';
import { NodeService, type NodeItem } from '@/lib/services/openflare';
import en from '@/messages/en.json';

let params = new URLSearchParams();
vi.mock('next/navigation', () => ({ useSearchParams: () => params }));
vi.mock('@/lib/services/openflare', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/services/openflare')>()),
  NodeService: {
    listNodePage: vi.fn(),
    listNodes: vi.fn(),
    getNode: vi.fn(),
    deleteNode: vi.fn(),
  },
}));
vi.mock('@/app/(main)/nodes/components/node-editor-dialog', () => ({
  NodeEditorDialog: () => null,
}));
vi.mock('@/app/(main)/nodes/components/nodes-table', () => ({
  NodesTable: ({
    nodes,
    onDelete,
  }: {
    nodes: NodeItem[];
    onDelete: (node: NodeItem) => void;
  }) => (
    <div>
      {nodes.map((node) => (
        <div key={node.id}>
          {node.name}
          <button onClick={() => onDelete(node)}>Delete {node.id}</button>
        </div>
      ))}
    </div>
  ),
}));
vi.mock('@/app/(main)/nodes/components/edge-node-detail', () => ({
  EdgeNodeDetail: ({ node }: { node: NodeItem }) => (
    <div>Node detail {node.id}</div>
  ),
}));
vi.mock('@/app/(main)/nodes/components/relay-node-detail', () => ({
  RelayNodeDetail: () => null,
}));
vi.mock('@/app/(main)/nodes/components/tunnel-node-detail', () => ({
  TunnelNodeDetail: () => null,
}));

function node(id: number): NodeItem {
  return { id, name: `Edge ${id}`, node_type: 'edge_node' } as NodeItem;
}
function renderPage(detail = false) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  const result = render(
    <NextIntlClientProvider locale='en' messages={en} timeZone='UTC'>
      <QueryClientProvider client={client}>
        {detail ? <NodeDetailPageClient /> : <NodesPageClient />}
      </QueryClientProvider>
    </NextIntlClientProvider>,
  );
  return { ...result, client };
}

beforeEach(() => {
  params = new URLSearchParams();
  vi.clearAllMocks();
  vi.mocked(NodeService.listNodePage).mockImplementation(
    async (query = {}) => ({
      page: query.page ?? 1,
      page_size: query.page_size ?? 20,
      total: 41,
      results: [node(query.page ?? 1)],
    }),
  );
});
afterEach(() => {
  cleanup();
});

it('fetches only visible pages and debounces server search while resetting the page', async () => {
  renderPage();
  await screen.findByText('Edge 1');
  fireEvent.click(screen.getByRole('button', { name: 'Next' }));
  await screen.findByText('Edge 2');
  expect(NodeService.listNodePage).toHaveBeenLastCalledWith(
    expect.objectContaining({ page: 2, page_size: 20 }),
  );
  expect(NodeService.listNodes).not.toHaveBeenCalled();
  fireEvent.change(screen.getByLabelText('Search nodes'), {
    target: { value: 'test' },
  });
  await waitFor(() =>
    expect(NodeService.listNodePage).toHaveBeenLastCalledWith(
      expect.objectContaining({ page: 1, search: 'test' }),
    ),
  );
});

it('refetches the previous valid page after deleting the only node on the last page', async () => {
  let deleted = false;
  vi.mocked(NodeService.listNodePage).mockImplementation(
    async (query = {}) => ({
      page: query.page ?? 1,
      page_size: 20,
      total: deleted ? 20 : 21,
      results:
        deleted && query.page === 2 ? [] : [node(query.page === 2 ? 21 : 1)],
    }),
  );
  vi.mocked(NodeService.deleteNode).mockImplementation(async () => {
    deleted = true;
  });
  renderPage();
  await screen.findByText('Edge 1');
  fireEvent.click(screen.getByRole('button', { name: 'Next' }));
  await screen.findByText('Edge 21');
  fireEvent.click(screen.getByRole('button', { name: 'Delete 21' }));
  fireEvent.click(await screen.findByRole('button', { name: 'Delete' }));
  await waitFor(() =>
    expect(NodeService.listNodePage).toHaveBeenLastCalledWith(
      expect.objectContaining({ page: 1 }),
    ),
  );
  await screen.findByText('Edge 1');
});

it('resets the page immediately when the URL node type changes', async () => {
  const page = renderPage();
  await screen.findByText('Edge 1');
  fireEvent.click(screen.getByRole('button', { name: 'Next' }));
  await screen.findByText('Edge 2');
  params = new URLSearchParams('filter=relay');
  page.rerender(
    <NextIntlClientProvider locale='en' messages={en} timeZone='UTC'>
      <QueryClientProvider client={page.client}>
        <NodesPageClient />
      </QueryClientProvider>
    </NextIntlClientProvider>,
  );
  await waitFor(() =>
    expect(NodeService.listNodePage).toHaveBeenLastCalledWith(
      expect.objectContaining({ page: 1, node_type: 'tunnel_relay' }),
    ),
  );
});

it('keeps page errors retryable through refresh', async () => {
  vi.mocked(NodeService.listNodePage).mockRejectedValueOnce(
    new Error('temporary failure'),
  );
  renderPage();
  await screen.findByText('temporary failure');
  fireEvent.click(screen.getByRole('button', { name: 'Refresh now' }));
  await screen.findByText('Edge 1');
});

it('fetches one detail and refetches it on shared node invalidation', async () => {
  params = new URLSearchParams('id=42');
  vi.mocked(NodeService.getNode).mockResolvedValue(node(42));
  const { client } = renderPage(true);
  await screen.findByText('Node detail 42');
  expect(NodeService.getNode).toHaveBeenCalledWith(42);
  expect(NodeService.listNodes).not.toHaveBeenCalled();
  await act(async () => {
    await client.invalidateQueries({ queryKey: ['openflare', 'nodes'] });
  });
  expect(NodeService.getNode).toHaveBeenCalledTimes(2);
});
