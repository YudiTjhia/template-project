'use client';

import PageContainer from '../page/PageContainer';
import EditComponent, { type EditComponentProps } from './EditComponent';

export type EditPageProps = EditComponentProps & {
  /** Maximum width of the page content. */
  maxWidth?: number | string;
  /** Horizontal and vertical page padding. */
  padding?: number | { xs?: number; sm?: number; md?: number; lg?: number };
};

export default function EditPage({
  maxWidth = 'none',
  padding = { xs: 2, md: 4 },
  ...editProps
}: EditPageProps) {
  return (
    <PageContainer maxWidth={maxWidth} padding={padding}>
      <EditComponent {...editProps} />
    </PageContainer>
  );
}
