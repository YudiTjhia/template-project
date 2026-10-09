'use client';

import PageContainer from '../page/PageContainer';
import ListComponent, { type ListComponentProps, type ListRow } from './ListComponent';

export type ListPageProps<T extends ListRow> = ListComponentProps<T> & {
  /** Maximum width of the page content. */
  maxWidth?: number | string;
  /** Horizontal and vertical page padding. */
  padding?: number | { xs?: number; sm?: number; md?: number; lg?: number };
};

export default function ListPage<T extends ListRow>({
  maxWidth = 1200,
  padding = { xs: 2, md: 4 },
  ...listProps
}: ListPageProps<T>) {
  return (
    <PageContainer maxWidth={maxWidth} padding={padding}>
      <ListComponent {...listProps} />
    </PageContainer>
  );
}
