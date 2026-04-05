import { mountSuspended } from '@nuxt/test-utils/runtime'
import DataTable from './DataTable.vue'

const columns = [
  { key: 'name', label: 'Name', sortable: true },
  { key: 'status', label: 'Status', sortable: true },
  { key: 'price', label: 'Price' },
]

const rows = [
  { id: '1', name: 'Widget', status: 'active', price: 29.99 },
  { id: '2', name: 'Gadget', status: 'draft', price: 49.99 },
]

describe('DataTable', () => {
  it('renders column headers', async () => {
    const wrapper = await mountSuspended(DataTable, { props: { columns, rows } })
    expect(wrapper.text()).toContain('Name')
    expect(wrapper.text()).toContain('Status')
    expect(wrapper.text()).toContain('Price')
  })

  it('renders row data', async () => {
    const wrapper = await mountSuspended(DataTable, { props: { columns, rows } })
    expect(wrapper.text()).toContain('Widget')
    expect(wrapper.text()).toContain('Gadget')
  })

  it('shows loading skeleton when loading', async () => {
    const wrapper = await mountSuspended(DataTable, { props: { columns, rows: [], loading: true } })
    expect(wrapper.findAll('.animate-pulse').length).toBeGreaterThan(0)
  })

  it('shows empty text when no rows and not loading', async () => {
    const wrapper = await mountSuspended(DataTable, { props: { columns, rows: [] } })
    expect(wrapper.text()).toContain('No data found')
  })

  it('shows custom empty text', async () => {
    const wrapper = await mountSuspended(DataTable, { props: { columns, rows: [], emptyText: 'Nothing here' } })
    expect(wrapper.text()).toContain('Nothing here')
  })

  it('clicking sortable header emits sort', async () => {
    const wrapper = await mountSuspended(DataTable, { props: { columns, rows } })
    const headers = wrapper.findAll('th')
    await headers[0].trigger('click') // Name is sortable
    expect(wrapper.emitted('sort')?.[0]).toEqual(['name', 'asc'])
  })

  it('clicking same header toggles direction', async () => {
    const wrapper = await mountSuspended(DataTable, { props: { columns, rows } })
    const headers = wrapper.findAll('th')
    await headers[0].trigger('click')
    await headers[0].trigger('click')
    expect(wrapper.emitted('sort')?.[1]).toEqual(['name', 'desc'])
  })

  it('non-sortable columns do not trigger sort', async () => {
    const wrapper = await mountSuspended(DataTable, { props: { columns, rows } })
    const headers = wrapper.findAll('th')
    await headers[2].trigger('click') // Price is not sortable
    expect(wrapper.emitted('sort')).toBeFalsy()
  })

  it('clicking row emits row-click', async () => {
    const wrapper = await mountSuspended(DataTable, { props: { columns, rows } })
    const tableRows = wrapper.findAll('tbody tr')
    await tableRows[0].trigger('click')
    expect(wrapper.emitted('row-click')?.[0][0]).toMatchObject({ name: 'Widget' })
  })

  it('pagination buttons emit update:page', async () => {
    const wrapper = await mountSuspended(DataTable, { props: { columns, rows, page: 2, totalPages: 3, total: 50 } })
    const prevBtn = wrapper.findAll('button').find(b => b.text() === 'Prev')!
    const nextBtn = wrapper.findAll('button').find(b => b.text() === 'Next')!
    await prevBtn.trigger('click')
    expect(wrapper.emitted('update:page')?.[0]).toEqual([1])
    await nextBtn.trigger('click')
    expect(wrapper.emitted('update:page')?.[1]).toEqual([3])
  })
})
